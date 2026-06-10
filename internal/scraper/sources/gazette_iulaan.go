package sources

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/ahmedazhan/azx-scrap-bot/internal/scraper/scrtypes"
	"github.com/ahmedazhan/azx-scrap-bot/internal/translate"
	"github.com/ahmedazhan/azx-scrap-bot/internal/util"
)

type GazetteIulaanSource struct{}

func (GazetteIulaanSource) Key() string { return "gazette-iulaan" }

func (GazetteIulaanSource) ListURLs(_ context.Context, maxPages int) ([]string, error) {
	if maxPages <= 0 {
		maxPages = 1
	}
	base := "https://gazette.gov.mv"
	urls := make([]string, 0, maxPages)
	for i := 1; i <= maxPages; i++ {
		if i == 1 {
			urls = append(urls, base+"/iulaan")
		} else {
			urls = append(urls, fmt.Sprintf("%s/iulaan?page=%d", base, i))
		}
	}
	return urls, nil
}

func (GazetteIulaanSource) ParseList(doc *goquery.Document, baseURL string) ([]scrtypes.ItemRef, error) {
	refs := []scrtypes.ItemRef{}
	doc.Find(".items-list .bordered.items").Each(func(_ int, s *goquery.Selection) {
		ref := scrtypes.ItemRef{}

		titleSel := s.Find("a.iulaan-title")
		title := util.TrimNonPrintable(strings.TrimSpace(titleSel.First().Text()))
		href, _ := titleSel.First().Attr("href")
		if href != "" {
			ref.URL = absolutize(baseURL, href)
		}
		ref.Title = title

		typeSel := s.Find("a.iulaan-type")
		ref.TypeSlug = slugFromHref(typeSel.First())
		_ = typeSel.First().Text()

		officeSel := s.Find("a.iulaan-office")
		ref.Office = util.CollapseWS(strings.TrimSpace(officeSel.First().Text()))

		info := s.Find(".info").First().Text()
		ref.DateText = util.CollapseWS(strings.TrimSpace(info))

		ref.ExternalID = extractIDFromURL(ref.URL)
		if ref.ExternalID == "" {
			ref.ExternalID = strings.TrimSpace(ref.Title)
		}
		if ref.ExternalID == "" {
			return
		}
		if ref.URL == "" {
			return
		}
		refs = append(refs, ref)
	})
	return refs, nil
}

func (GazetteIulaanSource) ParseDetail(doc *goquery.Document, baseURL string, ref scrtypes.ItemRef) (scrtypes.DetailFields, error) {
	d := scrtypes.DetailFields{
		ExternalID: ref.ExternalID,
		URL:        ref.URL,
		Title:      ref.Title,
		TypeSlug:   ref.TypeSlug,
		Office:     ref.Office,
	}

	if og := doc.Find("meta[property='og:url']").First(); og.Length() > 0 {
		if v, ok := og.Attr("content"); ok && v != "" {
			d.URL = absolutize(baseURL, v)
		}
	}
	if d.ExternalID == "" {
		d.ExternalID = extractIDFromURL(d.URL)
	}

	titleSel := doc.Find("div.iulaan.iulaan-title.center").First()
	if titleSel.Length() > 0 {
		d.Title = util.CollapseWS(strings.TrimSpace(titleSel.Text()))
	}

	if d.TypeSlug == "" {
		ts := doc.Find("a.iulaan-type").First()
		d.TypeSlug = slugFromHref(ts)
	}
	if entry, ok := translate.Lookup(d.TypeSlug); ok {
		d.TypeLabel = entry.English
		d.TypeDhivehi = entry.Dhivehi
	}

	if office := doc.Find("span.office-name").First(); office.Length() > 0 {
		d.Office = util.CollapseWS(strings.TrimSpace(office.Text()))
	}

	doc.Find(".additional-info .info").Each(func(_ int, row *goquery.Selection) {
		text := row.Text()
		if strings.HasPrefix(strings.TrimSpace(text), "ނަންބަރު:") {
			if en := row.Find("span.en").First(); en.Length() > 0 {
				d.ReferenceNo = util.CollapseWS(strings.TrimSpace(en.Text()))
			}
		}
		if strings.HasPrefix(strings.TrimSpace(text), "ޕަބްލިޝްކުރި ތާރީޚު:") {
			val := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "ޕަބްލިޝްކުރި ތާރީޚު:"))
			if t, ok := parseLooseDate(val); ok {
				d.PublishedAt = &t
			}
		}
		if strings.HasPrefix(strings.TrimSpace(text), "ޕަބްލިޝްކުރި ގަޑި:") {
			val := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "ޕަބްލިޝްކުރި ގަޑި:"))
			if t, ok := parseLooseDate(val); ok {
				_ = t
			}
		}
		if strings.HasPrefix(strings.TrimSpace(text), "ސުންގަޑި:") {
			val := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "ސުންގަޑި:"))
			if t, ok := parseLooseDate(val); ok {
				d.DeadlineAt = &t
			}
		}
	})

	details := doc.Find("#iulaan-view .iulaan-info .details")
	if details.Length() == 0 {
		details = doc.Find(".iulaan-info .details")
	}
	last := details.Last()
	if last.Length() > 0 {
		html, _ := last.Html()
		d.BodyHTML = util.TrimNonPrintable(html)
		d.BodyText = util.CollapseWS(strings.TrimSpace(last.Text()))
	}

	return d, nil
}

func absolutize(base, href string) string {
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if base == "" {
		return href
	}
	u, err := url.Parse(base)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	u.Path = path.Join(u.Path, ref.Path)
	u.RawQuery = ref.RawQuery
	return u.String()
}

func extractIDFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func slugFromHref(s *goquery.Selection) string {
	if s == nil {
		return ""
	}
	href, _ := s.Attr("href")
	if href == "" {
		return ""
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return strings.Trim(u.Query().Get("type"), "")
}

func parseLooseDate(s string) (time.Time, bool) {
	formats := []string{
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"2006/01/02",
		time.RFC3339,
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

var _ = path.Join
