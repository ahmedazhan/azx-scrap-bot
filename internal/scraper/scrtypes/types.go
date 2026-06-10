package scrtypes

import (
	"context"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ItemRef struct {
	ExternalID string
	URL        string
	TypeSlug   string
	Title      string
	Office     string
	DateText   string
}

type DetailFields struct {
	ExternalID  string
	URL         string
	Title       string
	TypeSlug    string
	TypeLabel   string
	TypeDhivehi string
	Office      string
	ReferenceNo string
	PublishedAt *time.Time
	DeadlineAt  *time.Time
	BodyText    string
	BodyHTML    string
}

type Source interface {
	Key() string
	ListURLs(ctx context.Context, maxPages int) ([]string, error)
	ParseList(doc *goquery.Document, baseURL string) ([]ItemRef, error)
	ParseDetail(doc *goquery.Document, baseURL string, ref ItemRef) (DetailFields, error)
}
