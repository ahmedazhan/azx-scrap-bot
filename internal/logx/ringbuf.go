package logx

import (
	"sync"
	"sync/atomic"
	"time"
)

type Entry struct {
	Seq   uint64
	Time  time.Time
	Level Level
	Msg   string
	Attrs map[string]any
}

type RingBuffer struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
	next    atomic.Uint64
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 500
	}
	return &RingBuffer{entries: make([]Entry, 0, capacity), cap: capacity}
}

func (r *RingBuffer) Push(e Entry) {
	e.Seq = r.next.Add(1)
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.entries) >= r.cap {
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, e)
}

func (r *RingBuffer) Snapshot() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Entry, len(r.entries))
	copy(out, r.entries)
	return out
}

func (r *RingBuffer) Recent(n int) []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n <= 0 || n > len(r.entries) {
		n = len(r.entries)
	}
	out := make([]Entry, n)
	copy(out, r.entries[len(r.entries)-n:])
	return out
}

func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}
