package scraper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"

	"github.com/ahmedazhan/azx-scrap-bot/internal/db"
	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper/scrtypes"
	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper/sources"
)

type Scheduler struct {
	db       *gorm.DB
	log      *slog.Logger
	client   *Client
	flagger  *Flagger
	sources  map[string]scrtypes.Source
	wg       sync.WaitGroup
	mu       sync.Mutex
	cancels  map[string]context.CancelFunc
	rootStop context.CancelFunc
	stopped  bool
}

func New(db *gorm.DB, log *slog.Logger, flagger *Flagger) *Scheduler {
	return &Scheduler{
		db:      db,
		log:     log,
		client:  NewClient(),
		flagger: flagger,
		sources: map[string]scrtypes.Source{
			"gazette-iulaan": sources.GazetteIulaanSource{},
		},
		cancels: make(map[string]context.CancelFunc),
	}
}

func (s *Scheduler) Start(parent context.Context) error {
	var rows []db.Source
	if err := s.db.Find(&rows).Error; err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(parent)
	s.mu.Lock()
	s.rootStop = cancel
	s.mu.Unlock()

	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		if _, ok := s.sources[row.Key]; !ok {
			s.log.Warn("no scraper implementation registered for source", "key", row.Key)
			continue
		}
		s.spawn(ctx, row)
	}
	return nil
}

func (s *Scheduler) spawn(ctx context.Context, row db.Source) {
	src := s.sources[row.Key]
	cctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancels[row.Key] = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		interval := time.Duration(row.IntervalSec) * time.Second
		if interval <= 0 {
			interval = 15 * time.Minute
		}
		s.runOnce(cctx, src)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-cctx.Done():
				return
			case <-t.C:
				s.runOnce(cctx, src)
			}
		}
	}()
}

func (s *Scheduler) runOnce(ctx context.Context, src scrtypes.Source) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("scrape cycle panic", "source", src.Key(), "panic", fmt.Sprint(r))
		}
	}()

	var sRow db.Source
	if err := s.db.Where("key = ?", src.Key()).First(&sRow).Error; err != nil {
		s.log.Warn("source row missing", "key", src.Key())
		return
	}

	maxPages := sRow.MaxPagesPerCycle
	if maxPages <= 0 {
		maxPages = 5
	}

	run := db.ScrapeRun{SourceID: sRow.ID, StartedAt: time.Now().UTC()}
	if err := s.db.Create(&run).Error; err != nil {
		s.log.Warn("create scrape run", "err", err)
	}

	urls, err := src.ListURLs(ctx, maxPages)
	if err != nil {
		s.log.Error("list urls", "source", src.Key(), "err", err)
		s.recordRunError(&run, err)
		return
	}

	refs := []scrtypes.ItemRef{}
	seen := map[string]struct{}{}
	for _, u := range urls {
		status, body, _, err := s.client.Fetch(ctx, u)
		if err != nil {
			s.log.Warn("list fetch failed", "url", u, "err", err)
			continue
		}
		if status < 200 || status >= 300 {
			s.log.Warn("list fetch bad status", "url", u, "status", status)
			continue
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			s.log.Warn("list parse", "url", u, "err", err)
			continue
		}
		parsed, err := src.ParseList(doc, baseFromURL(u))
		if err != nil {
			s.log.Warn("list parse", "url", u, "err", err)
			continue
		}
		for _, p := range parsed {
			if _, ok := seen[p.ExternalID]; ok {
				continue
			}
			seen[p.ExternalID] = struct{}{}
			refs = append(refs, p)
		}
	}

	pool := NewPool(sRow.Concurrency, s.client, s.log)
	results := pool.Run(ctx, refs, func(ctx context.Context, ref scrtypes.ItemRef) (scrtypes.DetailFields, error) {
		detailURL := ref.URL
		if detailURL == "" {
			return scrtypes.DetailFields{}, errors.New("empty detail url")
		}
		status, body, _, err := s.client.Fetch(ctx, detailURL)
		if err != nil {
			return scrtypes.DetailFields{}, err
		}
		if status < 200 || status >= 300 {
			return scrtypes.DetailFields{}, fmt.Errorf("detail status %d", status)
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return scrtypes.DetailFields{}, err
		}
		return src.ParseDetail(doc, baseFromURL(detailURL), ref)
	})

	newN, updN, errN := 0, 0, 0
	for _, r := range results {
		if r.Err != nil {
			errN++
			s.log.Warn("detail failed", "ref", r.Ref.ExternalID, "err", r.Err)
			continue
		}
		changed, isNew, err := s.upsertItem(sRow.ID, r.Ref, r.Detail)
		if err != nil {
			errN++
			s.log.Warn("upsert", "external_id", r.Ref.ExternalID, "err", err)
			continue
		}
		if isNew {
			newN++
		} else if changed {
			updN++
		}
	}

	finished := time.Now().UTC()
	run.FinishedAt = &finished
	run.NewItems = newN
	run.UpdatedItems = updN
	run.Fetched = len(results)
	run.Errors = errN
	s.db.Save(&run)

	s.db.Model(&db.Source{}).Where("id = ?", sRow.ID).Updates(map[string]any{
		"last_run_at": finished,
		"last_error":  "",
	})
	s.log.Info("scrape cycle done", "source", src.Key(), "new", newN, "updated", updN, "errors", errN)
}

func (s *Scheduler) recordRunError(run *db.ScrapeRun, err error) {
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	run.Errors = 1
	run.ErrorLog = err.Error()
	s.db.Save(run)
	s.db.Model(&db.Source{}).Where("id = ?", run.SourceID).Update("last_error", err.Error())
}

func (s *Scheduler) upsertItem(sourceID uint64, ref scrtypes.ItemRef, d scrtypes.DetailFields) (changed bool, isNew bool, err error) {
	var existing db.Item
	tx := s.db.Where("source_id = ? AND external_id = ?", sourceID, ref.ExternalID).First(&existing)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return false, false, tx.Error
	}
	now := time.Now().UTC()
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		item := db.Item{
			SourceID:    sourceID,
			ExternalID:  ref.ExternalID,
			URL:         d.URL,
			Title:       d.Title,
			TitleRaw:    ref.Title,
			TypeSlug:    d.TypeSlug,
			TypeLabel:   d.TypeLabel,
			TypeDhivehi: d.TypeDhivehi,
			Office:      d.Office,
			ReferenceNo: d.ReferenceNo,
			PublishedAt: d.PublishedAt,
			DeadlineAt:  d.DeadlineAt,
			BodyText:    d.BodyText,
			BodyHTML:    d.BodyHTML,
			ContentHash: hashContent(d),
			FetchedAt:   now,
		}
		if err := s.db.Create(&item).Error; err != nil {
			return false, false, err
		}
		if s.flagger != nil {
			s.flagger.Submit(&item)
		}
		return true, true, nil
	}

	newHash := hashContent(d)
	if existing.ContentHash == newHash {
		return false, false, nil
	}
	existing.URL = d.URL
	existing.Title = d.Title
	existing.TitleRaw = ref.Title
	existing.TypeSlug = d.TypeSlug
	existing.TypeLabel = d.TypeLabel
	existing.TypeDhivehi = d.TypeDhivehi
	existing.Office = d.Office
	existing.ReferenceNo = d.ReferenceNo
	existing.PublishedAt = d.PublishedAt
	existing.DeadlineAt = d.DeadlineAt
	existing.BodyText = d.BodyText
	existing.BodyHTML = d.BodyHTML
	existing.ContentHash = newHash
	existing.FetchedAt = now
	if err := s.db.Save(&existing).Error; err != nil {
		return false, false, err
	}
	if s.flagger != nil {
		s.flagger.Submit(&existing)
	}
	return true, false, nil
}

func (s *Scheduler) RunSourceNow(parent context.Context, key string) error {
	src, ok := s.sources[key]
	if !ok {
		return fmt.Errorf("unknown source %q", key)
	}
	s.mu.Lock()
	stopped := s.stopped
	s.mu.Unlock()
	if stopped {
		return errors.New("scheduler stopped")
	}
	go s.runOnce(parent, src)
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	for _, c := range s.cancels {
		c()
	}
	if s.rootStop != nil {
		s.rootStop()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func baseFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/")
}

func hashContent(d scrtypes.DetailFields) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%s\x00%v\x00%v\x00%s",
		d.Title, d.Office, d.ReferenceNo,
		d.TypeSlug, d.TypeDhivehi,
		ptrTime(d.PublishedAt), ptrTime(d.DeadlineAt),
		d.BodyText,
	)
	return hex.EncodeToString(h.Sum(nil))
}

func ptrTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
