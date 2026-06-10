package scraper

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper/scrtypes"
)

type Result struct {
	Ref    scrtypes.ItemRef
	Detail scrtypes.DetailFields
	Err    error
}

type Pool struct {
	Concurrency int
	Client      *Client
	Log         *slog.Logger
}

func NewPool(c int, client *Client, log *slog.Logger) *Pool {
	if c <= 0 {
		c = 4
	}
	return &Pool{Concurrency: c, Client: client, Log: log}
}

func (p *Pool) Run(ctx context.Context, items []scrtypes.ItemRef, fetch func(ctx context.Context, ref scrtypes.ItemRef) (scrtypes.DetailFields, error)) []Result {
	if len(items) == 0 {
		return nil
	}
	results := make([]Result, len(items))
	sem := make(chan struct{}, p.Concurrency)
	var wg sync.WaitGroup
	for i, it := range items {
		select {
		case <-ctx.Done():
			results[i] = Result{Ref: it, Err: ctx.Err()}
			continue
		default:
		}
		i, it := i, it
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			detail, err := fetch(ctx, it)
			results[i] = Result{Ref: it, Detail: detail, Err: err}
		}()
	}
	wg.Wait()
	return results
}
