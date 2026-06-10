package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

type Dispatcher struct {
	db     *gorm.DB
	log    *slog.Logger
	queue  chan hitEvent
	mu     sync.RWMutex
	bot    *Bot
	cfg    db.TelegramConfig
	stopCh chan struct{}
	wg     sync.WaitGroup
}

type hitEvent struct {
	ItemID uint64
	RuleID uint64
}

func NewDispatcher(database *gorm.DB, log *slog.Logger) *Dispatcher {
	d := &Dispatcher{
		db:     database,
		log:    log,
		queue:  make(chan hitEvent, 512),
		stopCh: make(chan struct{}),
	}
	d.reloadConfig()
	return d
}

func (d *Dispatcher) reloadConfig() {
	var c db.TelegramConfig
	err := d.db.First(&c, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c = db.TelegramConfig{ID: 1, ThrottleMs: 250, Enabled: false}
			if e := d.db.Create(&c).Error; e == nil {
				d.log.Info("telegram config bootstrapped")
			}
		} else {
			d.log.Warn("telegram config load", "err", err)
			return
		}
	}
	d.mu.Lock()
	d.cfg = c
	if c.BotToken != "" {
		throttle := time.Duration(c.ThrottleMs) * time.Millisecond
		d.bot = NewBot(c.BotToken, throttle)
	} else {
		d.bot = nil
	}
	d.mu.Unlock()
}

func (d *Dispatcher) Start(parent context.Context) error {
	d.wg.Add(1)
	go d.loop(parent)
	return nil
}

func (d *Dispatcher) Stop() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
	d.wg.Wait()
}

func (d *Dispatcher) loop(ctx context.Context) {
	defer d.wg.Done()
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case ev := <-d.queue:
			d.handle(ctx, ev)
		case <-t.C:
			d.reloadConfig()
		}
	}
}

func (d *Dispatcher) handle(ctx context.Context, ev hitEvent) {
	d.mu.RLock()
	cfg := d.cfg
	bot := d.bot
	d.mu.RUnlock()

	if !cfg.Enabled || bot == nil {
		return
	}

	var item db.Item
	if err := d.db.First(&item, ev.ItemID).Error; err != nil {
		return
	}
	var rule db.FlagRule
	_ = d.db.First(&rule, ev.RuleID).Error

	if cfg.NotifyOnFlagOnly && ev.RuleID == 0 {
		return
	}

	var subs []db.TelegramSubscriber
	if err := d.db.Where("enabled = ?", true).Find(&subs).Error; err != nil {
		return
	}
	if len(subs) == 0 {
		return
	}

	msg := formatMessage(item, rule)

	for _, s := range subs {
		if err := bot.Send(ctx, s.ChatID, msg); err != nil {
			d.log.Warn("telegram send failed", "chat", s.ChatID, "err", err)
			continue
		}
	}
	if ev.RuleID != 0 {
		d.db.Model(&db.FlagHit{}).Where("item_id = ? AND rule_id = ?", ev.ItemID, ev.RuleID).Update("notified", true)
	}
}

func formatMessage(item db.Item, rule db.FlagRule) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>\n", html.EscapeString(item.Title))
	if item.TypeLabel != "" || item.TypeDhivehi != "" {
		fmt.Fprintf(&b, "Type: %s / %s\n", html.EscapeString(item.TypeLabel), html.EscapeString(item.TypeDhivehi))
	}
	if item.Office != "" {
		fmt.Fprintf(&b, "Office: %s\n", html.EscapeString(item.Office))
	}
	if item.ReferenceNo != "" {
		fmt.Fprintf(&b, "Ref: %s\n", html.EscapeString(item.ReferenceNo))
	}
	if item.DeadlineAt != nil {
		fmt.Fprintf(&b, "Deadline: %s\n", item.DeadlineAt.UTC().Format("2006-01-02"))
	}
	if rule.Name != "" {
		fmt.Fprintf(&b, "Matched rule: %s\n", html.EscapeString(rule.Name))
	}
	if item.URL != "" {
		fmt.Fprintf(&b, "\n<a href=\"%s\">Open</a>", html.EscapeString(item.URL))
	}
	return b.String()
}

func (d *Dispatcher) Notify(itemID uint64, ruleID uint64) {
	select {
	case d.queue <- hitEvent{ItemID: itemID, RuleID: ruleID}:
	default:
		d.log.Warn("dispatcher queue full", "item", itemID, "rule", ruleID)
	}
}
