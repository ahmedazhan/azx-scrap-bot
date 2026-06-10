package logx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type bufioHandler struct {
	w   *bufio.Writer
	mu  *sync.Mutex
	enc *jsonEncoder
}

func newBufioHandler(w io.Writer) *bufioHandler {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &bufioHandler{
		w:   bw,
		mu:  &sync.Mutex{},
		enc: newJSONEncoder(),
	}
}

func (h *bufioHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *bufioHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enc.reset()
	h.enc.addTime(r.Time)
	h.enc.addLevel(r.Level)
	h.enc.addString("msg", r.Message)
	r.Attrs(func(a slog.Attr) bool {
		h.enc.addAttr(a)
		return true
	})
	line, err := h.enc.bytes()
	if err != nil {
		return err
	}
	if _, err := h.w.Write(line); err != nil {
		return err
	}
	if _, err := h.w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func (h *bufioHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *bufioHandler) WithGroup(_ string) slog.Handler      { return h }

type jsonEncoder struct {
	buf []byte
}

func newJSONEncoder() *jsonEncoder { return &jsonEncoder{} }

func (e *jsonEncoder) reset() { e.buf = e.buf[:0] }

func (e *jsonEncoder) addTime(t interface{ Format(string) string }) {
	e.addString("time", t.Format("2006-01-02T15:04:05.000Z07:00"))
}

func (e *jsonEncoder) addLevel(l slog.Level) {
	e.addString("level", l.String())
}

func (e *jsonEncoder) addString(k, v string) {
	e.buf = append(e.buf, '"')
	e.buf = appendEscapedString(e.buf, k)
	e.buf = append(e.buf, `":"`...)
	e.buf = appendEscapedString(e.buf, v)
	e.buf = append(e.buf, '"')
}

func (e *jsonEncoder) addAttr(a slog.Attr) {
	e.buf = append(e.buf, ',')
	e.buf = append(e.buf, '"')
	e.buf = appendEscapedString(e.buf, a.Key)
	e.buf = append(e.buf, `":`...)
	e.appendValue(a.Value)
}

func (e *jsonEncoder) appendValue(v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		e.buf = append(e.buf, '"')
		e.buf = appendEscapedString(e.buf, v.String())
		e.buf = append(e.buf, '"')
	case slog.KindInt64:
		e.buf = fmt.Appendf(e.buf, "%d", v.Int64())
	case slog.KindFloat64:
		e.buf = fmt.Appendf(e.buf, "%g", v.Float64())
	case slog.KindBool:
		if v.Bool() {
			e.buf = append(e.buf, "true"...)
		} else {
			e.buf = append(e.buf, "false"...)
		}
	case slog.KindDuration:
		e.buf = append(e.buf, '"')
		e.buf = appendEscapedString(e.buf, v.Duration().String())
		e.buf = append(e.buf, '"')
	case slog.KindTime:
		e.buf = append(e.buf, '"')
		e.buf = appendEscapedString(e.buf, v.Time().Format("2006-01-02T15:04:05.000Z07:00"))
		e.buf = append(e.buf, '"')
	case slog.KindGroup:
		e.buf = append(e.buf, '{')
		first := true
		for _, a := range v.Group() {
			if !first {
				e.buf = append(e.buf, ',')
			}
			first = false
			e.buf = append(e.buf, '"')
			e.buf = appendEscapedString(e.buf, a.Key)
			e.buf = append(e.buf, `":`...)
			e.appendValue(a.Value)
		}
		e.buf = append(e.buf, '}')
	default:
		e.buf = append(e.buf, '"')
		e.buf = appendEscapedString(e.buf, fmt.Sprint(v.Any()))
		e.buf = append(e.buf, '"')
	}
}

func (e *jsonEncoder) bytes() ([]byte, error) {
	out := make([]byte, 0, len(e.buf)+1)
	out = append(out, '{')
	out = append(out, e.buf...)
	out = append(out, '}')
	return out, nil
}

func appendEscapedString(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '\\':
			dst = append(dst, '\\', c)
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if c < 0x20 {
				dst = fmt.Appendf(dst, "\\u%04x", c)
			} else {
				dst = append(dst, c)
			}
		}
	}
	return dst
}
