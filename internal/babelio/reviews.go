package babelio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Review struct {
	Author    string `json:"author,omitempty"`
	AuthorURL string `json:"author_url,omitempty"`
	Rating    int    `json:"rating,omitempty"` // 0-5
	Date      string `json:"date,omitempty"`   // raw French date (e.g. "17 février 2022")
	DateISO   string `json:"date_iso,omitempty"`
	Body      string `json:"body,omitempty"`
}

type ReviewsPage struct {
	Reviews     []Review `json:"reviews"`
	Page        int      `json:"page"`
	TotalPages  int      `json:"total_pages,omitempty"`
	HasNextPage bool     `json:"has_next_page"`
}

var (
	frenchDateRE = regexp.MustCompile(`(\d{1,2})\s+(janvier|février|mars|avril|mai|juin|juillet|août|septembre|octobre|novembre|décembre)\s+(\d{4})`)
	monthMap     = map[string]string{
		"janvier": "01", "février": "02", "mars": "03", "avril": "04",
		"mai": "05", "juin": "06", "juillet": "07", "août": "08",
		"septembre": "09", "octobre": "10", "novembre": "11", "décembre": "12",
	}
)

// ParseReviews reads a /livres/SLUG/ID/critiques page (or the book detail
// page, which also contains reviews) and returns the list of critiques.
func ParseReviews(doc *goquery.Document) ReviewsPage {
	var page ReviewsPage
	doc.Find("div.post.post_con").Each(func(_ int, post *goquery.Selection) {
		var r Review
		header := post.Find("div.entete_critique").First()
		// Author: the first /monprofil.php link with non-empty text (the first
		// one is usually wrapping just the avatar image).
		header.Find(`a[href*="/monprofil.php"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
			txt := strings.TrimSpace(a.Text())
			if txt == "" {
				return true
			}
			r.Author = txt
			if href, ok := a.Attr("href"); ok {
				r.AuthorURL = absoluteURL(href)
			}
			return false
		})
		// Rating
		if rv, ok := post.Find("div.rateit").First().Attr("data-rateit-value"); ok {
			if f, err := strconv.ParseFloat(rv, 64); err == nil {
				r.Rating = int(f + 0.5)
			}
		}
		// Date
		if m := frenchDateRE.FindStringSubmatch(header.Text()); m != nil {
			r.Date = m[0]
			day := m[1]
			if len(day) == 1 {
				day = "0" + day
			}
			if mn, ok := monthMap[strings.ToLower(m[2])]; ok {
				r.DateISO = fmt.Sprintf("%s-%s-%s", m[3], mn, day)
			}
		}
		// Body
		body := post.Find("div.cri_corps_critique").First()
		r.Body = cleanText(body.Text())
		page.Reviews = append(page.Reviews, r)
	})

	// Pagination
	maxPage := 1
	doc.Find(`a[href*="pageN="]`).Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		if m := pageNRE.FindStringSubmatch(href); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > maxPage {
				maxPage = n
			}
		}
	})
	page.TotalPages = maxPage
	return page
}
