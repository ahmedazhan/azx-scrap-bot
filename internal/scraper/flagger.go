package scraper

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
)

type Flagger struct {
	db        *gorm.DB
	log       *slog.Logger
	queue     chan uint64
	rules     atomic.Pointer[[]compiledRule]
	stopCh    chan struct{}
	notifyHit func(itemID uint64, ruleID uint64)
}

type compiledRule struct {
	Rule    db.FlagRule
	regex   *regexp.Regexp
	pattern string
}

func NewFlagger(database *gorm.DB, log *slog.Logger, notify func(itemID uint64, ruleID uint64)) *Flagger {
	f := &Flagger{
		db:        database,
		log:       log,
		queue:     make(chan uint64, 1024),
		stopCh:    make(chan struct{}),
		notifyHit: notify,
	}
	f.reload()
	return f
}

func (f *Flagger) reload() {
	var rows []db.FlagRule
	if err := f.db.Where("enabled = ?", true).Find(&rows).Error; err != nil {
		f.log.Warn("flagger: load rules", "err", err)
		return
	}
	out := make([]compiledRule, 0, len(rows))
	for _, r := range rows {
		var re *regexp.Regexp
		pattern := r.Pattern
		if r.IsRegex {
			pattern := r.Pattern
			if !r.CaseSensitive {
				pattern = "(?i)" + pattern
			}
			re, _ = regexp.Compile(pattern)
			if re == nil {
				f.log.Warn("flagger: invalid regex", "id", r.ID, "pattern", r.Pattern)
				continue
			}
		} else {
			pattern = r.Pattern
			if !r.CaseSensitive {
				pattern = strings.ToLower(pattern)
			}
		}
		out = append(out, compiledRule{Rule: r, regex: re, pattern: pattern})
	}
	f.rules.Store(&out)
}

func (f *Flagger) ReloadLoop(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-f.stopCh:
			return
		case <-t.C:
			f.reload()
		}
	}
}

func (f *Flagger) Start(parent context.Context) error {
	go f.worker(parent)
	go f.ReloadLoop(parent)
	return nil
}

func (f *Flagger) Stop() {
	select {
	case <-f.stopCh:
	default:
		close(f.stopCh)
	}
}

func (f *Flagger) Submit(item *db.Item) {
	if item == nil {
		return
	}
	select {
	case f.queue <- item.ID:
	default:
		f.log.Warn("flagger queue full, dropping item", "id", item.ID)
	}
}

func (f *Flagger) SubmitByID(item *db.Item) {
	f.Submit(item)
}

func (f *Flagger) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-f.stopCh:
			return
		case id := <-f.queue:
			f.process(ctx, id)
		}
	}
}

func (f *Flagger) process(ctx context.Context, id uint64) {
	var item db.Item
	if err := f.db.First(&item, id).Error; err != nil {
		return
	}
	rules := f.Rules()
	if len(rules) == 0 {
		return
	}
	for _, cr := range rules {
		field := strings.ToLower(cr.Rule.MatchIn)
		var haystack string
		switch field {
		case "title":
			haystack = item.Title
		case "title_raw":
			haystack = item.TitleRaw
		case "office":
			haystack = item.Office
		case "refno", "reference_no":
			haystack = item.ReferenceNo
		case "body", "body_text", "":
			haystack = item.BodyText
		default:
			haystack = item.BodyText
		}
		if !cr.Rule.CaseSensitive {
			haystack = strings.ToLower(haystack)
		}
		snippet, ok := matchRule(cr, haystack)
		if !ok {
			continue
		}
		hit := db.FlagHit{ItemID: item.ID, RuleID: cr.Rule.ID, Snippet: snippet, Notified: false, CreatedAt: time.Now().UTC()}
		err := f.db.Where("item_id = ? AND rule_id = ?", item.ID, cr.Rule.ID).FirstOrCreate(&hit).Error
		if err != nil {
			f.log.Warn("flagger: save hit", "err", err)
			continue
		}
		if f.notifyHit != nil {
			f.notifyHit(item.ID, cr.Rule.ID)
		}
	}
}

func (f *Flagger) Rules() []compiledRule {
	p := f.rules.Load()
	if p == nil {
		return nil
	}
	return *p
}

func matchRule(cr compiledRule, haystack string) (string, bool) {
	if cr.Rule.IsRegex && cr.regex != nil {
		loc := cr.regex.FindStringIndex(haystack)
		if loc == nil {
			return "", false
		}
		return haystack[loc[0]:loc[1]], true
	}
	idx := strings.Index(haystack, cr.pattern)
	if idx < 0 {
		return "", false
	}
	return haystack[idx : idx+len(cr.pattern)], true
}

func (f *Flagger) TestRule(pattern string, isRegex, caseSensitive bool, matchIn string, itemIDs []uint64) (int, []testSample) {
	if matchIn == "" {
		matchIn = "body"
	}
	samples := []testSample{}
	matches := 0
	var items []db.Item
	q := f.db.Model(&db.Item{})
	if len(itemIDs) > 0 {
		q = q.Where("id IN ?", itemIDs)
	}
	if err := q.Find(&items).Error; err != nil {
		return 0, samples
	}
	cr := compiledRule{Rule: db.FlagRule{Pattern: pattern, IsRegex: isRegex, CaseSensitive: caseSensitive, MatchIn: matchIn}}
	if isRegex {
		cr.regex = regexp.MustCompile(pattern)
	} else {
		cr.pattern = pattern
		if !caseSensitive {
			cr.pattern = strings.ToLower(cr.pattern)
		}
	}
	for _, it := range items {
		var haystack string
		switch strings.ToLower(matchIn) {
		case "title":
			haystack = it.Title
		case "title_raw":
			haystack = it.TitleRaw
		case "office":
			haystack = it.Office
		case "refno", "reference_no":
			haystack = it.ReferenceNo
		default:
			haystack = it.BodyText
		}
		probe := haystack
		if !caseSensitive {
			probe = strings.ToLower(probe)
		}
		snippet, ok := matchRule(cr, probe)
		if ok {
			matches++
			if len(samples) < 5 {
				samples = append(samples, testSample{ItemID: it.ID, Snippet: snippet})
			}
		}
	}
	return matches, samples
}

type testSample struct {
	ItemID  uint64 `json:"item_id"`
	Snippet string `json:"snippet"`
}
