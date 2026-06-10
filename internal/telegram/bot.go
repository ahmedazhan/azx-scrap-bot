package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Bot struct {
	Token    string
	HTTP     *http.Client
	Throttle time.Duration
}

func NewBot(token string, throttle time.Duration) *Bot {
	if throttle <= 0 {
		throttle = 250 * time.Millisecond
	}
	return &Bot{
		Token:    token,
		Throttle: throttle,
		HTTP:     &http.Client{Timeout: 20 * time.Second},
	}
}

type sendMessageReq struct {
	ChatID                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type apiResp struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	ErrorCode   int    `json:"error_code"`
	Parameters  *struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

func (b *Bot) Send(ctx context.Context, chatID int64, text string) error {
	if b.Token == "" {
		return errors.New("bot token empty")
	}
	body := sendMessageReq{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}
	buf, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(b.Throttle)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := b.HTTP.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(backoff(attempt))
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var ar apiResp
		_ = json.Unmarshal(raw, &ar)
		if resp.StatusCode == http.StatusTooManyRequests && ar.Parameters != nil && ar.Parameters.RetryAfter > 0 {
			wait := time.Duration(ar.Parameters.RetryAfter) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			lastErr = fmt.Errorf("rate limited")
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 && ar.OK {
			return nil
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("telegram 5xx: %s", ar.Description)
			time.Sleep(backoff(attempt))
			continue
		}
		return fmt.Errorf("telegram send: %s", ar.Description)
	}
	if lastErr == nil {
		lastErr = errors.New("max retries")
	}
	return lastErr
}

func backoff(a int) time.Duration {
	return time.Duration(500*(1<<uint(a))) * time.Millisecond
}

func ParseChatID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
