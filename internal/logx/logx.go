package logx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Options struct {
	Dir         string
	Level       slog.Level
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	Compress    bool
	RingSize    int
	BufferSize  int
	FlushPeriod time.Duration
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN", "warning":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	}
	return slog.LevelInfo
}

type Logger struct {
	Log     *slog.Logger
	Ring    *RingBuffer
	writer  *bufio.Writer
	rotator *lumberjack.Logger
	mu      sync.Mutex
	stop    chan struct{}
	wg      sync.WaitGroup
	closed  bool
}

func New(opts Options) (*Logger, error) {
	if opts.Dir == "" {
		opts.Dir = "./logs"
	}
	if opts.RingSize <= 0 {
		opts.RingSize = 500
	}
	if opts.BufferSize <= 0 {
		opts.BufferSize = 64 * 1024
	}
	if opts.FlushPeriod <= 0 {
		opts.FlushPeriod = time.Second
	}
	if opts.MaxSizeMB <= 0 {
		opts.MaxSizeMB = 50
	}
	if opts.MaxBackups <= 0 {
		opts.MaxBackups = 5
	}
	if opts.MaxAgeDays <= 0 {
		opts.MaxAgeDays = 30
	}

	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return nil, err
	}

	rotator := &lumberjack.Logger{
		Filename:   opts.Dir + "/azx-scrap-bot.log",
		MaxSize:    opts.MaxSizeMB,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAgeDays,
		Compress:   opts.Compress,
		LocalTime:  true,
	}

	bw := bufio.NewWriterSize(rotator, opts.BufferSize)

	ring := NewRingBuffer(opts.RingSize)

	fileHandler := slog.NewJSONHandler(bw, &slog.HandlerOptions{
		Level: opts.Level,
	})
	consoleHandler := newColorHandler(os.Stderr, &slog.HandlerOptions{
		Level: opts.Level,
	})

	tee := &teeHandler{handlers: []slog.Handler{consoleHandler, &ringHandler{ring: ring}, fileHandler}}
	log := slog.New(tee)

	l := &Logger{
		Log:     log,
		Ring:    ring,
		writer:  bw,
		rotator: rotator,
		stop:    make(chan struct{}),
	}

	l.wg.Add(1)
	go l.flushLoop(opts.FlushPeriod)

	return l, nil
}

func (l *Logger) flushLoop(period time.Duration) {
	defer l.wg.Done()
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-t.C:
			_ = l.Flush()
		}
	}
}

func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	return l.writer.Flush()
}

func (l *Logger) Sync() error {
	return l.Shutdown(context.Background())
}

func (l *Logger) Shutdown(ctx context.Context) error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	close(l.stop)
	l.mu.Unlock()

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.writer.Flush()
	return l.rotator.Close()
}

type teeHandler struct {
	handlers []slog.Handler
}

func (t *teeHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (t *teeHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range t.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, 0, len(t.handlers))
	for _, h := range t.handlers {
		out = append(out, h.WithAttrs(attrs))
	}
	return &teeHandler{handlers: out}
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, 0, len(t.handlers))
	for _, h := range t.handlers {
		out = append(out, h.WithGroup(name))
	}
	return &teeHandler{handlers: out}
}

type ringHandler struct {
	ring *RingBuffer
}

func (h *ringHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *ringHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = valueOf(a.Value)
		return true
	})
	h.ring.Push(Entry{
		Time:  r.Time,
		Level: r.Level,
		Msg:   r.Message,
		Attrs: attrs,
	})
	return nil
}

func (h *ringHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *ringHandler) WithGroup(_ string) slog.Handler      { return h }

type colorHandler struct {
	w     io.Writer
	opts  *slog.HandlerOptions
	mu    *sync.Mutex
	attrs []slog.Attr
	group string
}

func newColorHandler(w io.Writer, opts *slog.HandlerOptions) *colorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &colorHandler{w: w, opts: opts, mu: &sync.Mutex{}}
}

const (
	colReset  = "\x1b[0m"
	colGray   = "\x1b[90m"
	colCyan   = "\x1b[36m"
	colGreen  = "\x1b[32m"
	colYellow = "\x1b[33m"
	colRed    = "\x1b[31m"
	colBlue   = "\x1b[34m"
	colBold   = "\x1b[1m"
)

func levelColor(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return colRed
	case l >= slog.LevelWarn:
		return colYellow
	case l >= slog.LevelInfo:
		return colGreen
	default:
		return colCyan
	}
}

func (h *colorHandler) Enabled(_ context.Context, l slog.Level) bool {
	if h.opts.Level == nil {
		return true
	}
	return l >= h.opts.Level.Level()
}

func (h *colorHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ts := r.Time.Format("15:04:05.000")
	level := levelColor(r.Level)
	fmt.Fprintf(h.w, "%s%s %s%-5s%s %s%s%s",
		colGray, ts,
		level, r.Level.String(), colReset,
		colBold, r.Message, colReset,
	)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(h.w, " %s%s=%v%s", colBlue, a.Key, valueOf(a.Value), colReset)
		return true
	})
	fmt.Fprintln(h.w)
	return nil
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append([]slog.Attr{}, h.attrs...)
	merged = append(merged, attrs...)
	return &colorHandler{w: h.w, opts: h.opts, mu: h.mu, attrs: merged, group: h.group}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &colorHandler{w: h.w, opts: h.opts, mu: h.mu, attrs: h.attrs, group: name}
}

func valueOf(v slog.Value) any {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindTime:
		return v.Time()
	case slog.KindGroup:
		return v.Group()
	case slog.KindAny:
		fallthrough
	default:
		return v.Any()
	}
}
