package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/126.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
}

type Client struct {
	http   *req.Client
	muMap  sync.Mutex
	hostAt map[string]time.Time
	rng    *rand.Rand
	rngMu  sync.Mutex
}

func NewClient() *Client {
	c := req.NewClient().
		SetTimeout(30 * time.Second).
		SetRedirectPolicy(req.NoRedirectPolicy()).
		SetCommonHeaders(map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language": "en-US,en;q=0.9,dv;q=0.8",
		})
	return &Client{
		http:   c,
		hostAt: make(map[string]time.Time),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Client) randomUA() string {
	c.rngMu.Lock()
	defer c.rngMu.Unlock()
	return userAgents[c.rng.Intn(len(userAgents))]
}

func (c *Client) jitter(min, max time.Duration) time.Duration {
	c.rngMu.Lock()
	defer c.rngMu.Unlock()
	if max <= min {
		return min
	}
	return min + time.Duration(c.rng.Int63n(int64(max-min)))
}

func (c *Client) pace(ctx context.Context, rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	host := u.Host
	c.muMap.Lock()
	last := c.hostAt[host]
	now := time.Now()
	if last.IsZero() {
		c.hostAt[host] = now
		c.muMap.Unlock()
		return
	}
	c.muMap.Unlock()
	wait := c.jitter(250*time.Millisecond, 500*time.Millisecond) - time.Since(last)
	if wait < 0 {
		wait = 0
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
	c.muMap.Lock()
	c.hostAt[host] = time.Now()
	c.muMap.Unlock()
}

func (c *Client) Fetch(ctx context.Context, rawURL string) (int, []byte, time.Duration, error) {
	c.pace(ctx, rawURL)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		start := time.Now()
		r := c.http.R().
			SetContext(ctx).
			SetHeader("User-Agent", c.randomUA()).
			SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8").
			SetHeader("Accept-Language", "en-US,en;q=0.9,dv;q=0.8")
		resp, err := r.Get(rawURL)
		dur := time.Since(start)
		if err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return 0, nil, dur, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server status %d", resp.StatusCode)
			select {
			case <-ctx.Done():
				return resp.StatusCode, nil, dur, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return resp.StatusCode, nil, dur, err
		}
		return resp.StatusCode, body, dur, nil
	}
	if lastErr == nil {
		lastErr = errors.New("max retries")
	}
	return 0, nil, 0, lastErr
}

func backoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	return base * time.Duration(1<<uint(attempt))
}

func HostFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}

func NormalizeBase(raw string) string {
	raw = strings.TrimRight(raw, "/")
	if !strings.HasPrefix(raw, "http") {
		raw = "https://" + raw
	}
	return raw
}

var _ = http.MethodGet
