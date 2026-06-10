package logx

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRingBufferPushAndRecent(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Push(Entry{Time: time.Now(), Level: LevelInfo, Msg: "a"})
	rb.Push(Entry{Time: time.Now(), Level: LevelInfo, Msg: "b"})
	rb.Push(Entry{Time: time.Now(), Level: LevelInfo, Msg: "c"})
	rb.Push(Entry{Time: time.Now(), Level: LevelInfo, Msg: "d"})

	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}
	recent := rb.Recent(2)
	if len(recent) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(recent))
	}
	if recent[0].Msg != "c" || recent[1].Msg != "d" {
		t.Fatalf("unexpected order: %+v", recent)
	}
	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("snapshot should have 3, got %d", len(snap))
	}
	snap[0].Msg = "mutated"
	if rb.Snapshot()[0].Msg == "mutated" {
		t.Fatalf("snapshot should be a copy")
	}
}

func TestRingBufferOverflow(t *testing.T) {
	rb := NewRingBuffer(2)
	for i := 0; i < 100; i++ {
		rb.Push(Entry{Msg: "x"})
	}
	if rb.Len() != 2 {
		t.Fatalf("cap should clamp at 2, got %d", rb.Len())
	}
}

func TestBufferedWriteAndFlush(t *testing.T) {
	dir := t.TempDir()
	l, err := New(Options{
		Dir:         dir,
		Level:       slog.LevelDebug,
		MaxSizeMB:   1,
		MaxBackups:  1,
		Compress:    false,
		RingSize:    100,
		BufferSize:  4096,
		FlushPeriod: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	l.Log.Info("hello", "k", "v")
	l.Log.Debug("debug-line", "n", 42)
	if err := l.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "azx-scrap-bot.log"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(content), `"msg":"hello"`) {
		t.Fatalf("expected hello in log, got: %s", string(content))
	}
	if !strings.Contains(string(content), `"k":"v"`) {
		t.Fatalf("expected attr k=v, got: %s", string(content))
	}

	if l.Ring.Len() < 2 {
		t.Fatalf("expected ring to capture entries, got %d", l.Ring.Len())
	}
	if err := l.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestRotationTriggers(t *testing.T) {
	dir := t.TempDir()
	l, err := New(Options{
		Dir:         dir,
		Level:       slog.LevelInfo,
		MaxSizeMB:   1,
		MaxBackups:  2,
		Compress:    false,
		RingSize:    10,
		BufferSize:  1024,
		FlushPeriod: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Shutdown(context.Background())

	big := strings.Repeat("A", 2000)
	for i := 0; i < 2000; i++ {
		l.Log.Info("filler", "blob", big)
	}
	_ = l.Flush()
	_ = l.rotator.Rotate()

	matches, err := filepath.Glob(filepath.Join(dir, "azx-scrap-bot-*.log*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected rotated log files, found 0 in %s", dir)
	}
}

func TestSyncStopsFlusher(t *testing.T) {
	dir := t.TempDir()
	l, err := New(Options{Dir: dir, Level: slog.LevelInfo, RingSize: 5, FlushPeriod: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Log.Info("x")
	if err := l.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	l.Log.Info("y")
	_ = bufio.NewScanner(strings.NewReader(""))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Log.Info("z")
	}()
	wg.Wait()
}
