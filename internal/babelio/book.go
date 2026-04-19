package babelio

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type BookDetail struct {
	BookID    string   `json:"book_id"`
	Title     string   `json:"title"`
	Author    string   `json:"author,omitempty"`
	AuthorID  string   `json:"author_id,omitempty"`
	URL       string   `json:"url"`
	Synopsis  string   `json:"synopsis,omitempty"`
	Publisher string   `json:"publisher,omitempty"`
	DatePub   string   `json:"date_published,omitempty"`
	Pages     int      `json:"pages,omitempty"`
	AvgRating float64  `json:"avg_rating,omitempty"`
	NbRatings int      `json:"nb_ratings,omitempty"`
	NbReviews int      `json:"nb_reviews,omitempty"`
	Genres    []string `json:"genres,omitempty"`
}

var (
	nbPagesRE  = regexp.MustCompile(`(\d+)\s*pages?`)
	datePubRE  = regexp.MustCompile(`(?i)date de parution[^:]*:\s*([^\n]+)`)
	bookPathRE = regexp.MustCompile(`/livres/[^/]+/(\d+)`)
)

// ResolveBookPath extracts a canonical `/livres/SLUG/ID` path from a variety
// of user-supplied inputs (full URL, path, or "slug/id"). Returns an error if
// no book ID can be located.
func ResolveBookPath(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", errors.New("empty book reference")
	}
	// Strip protocol/host if present
	s = strings.TrimPrefix(s, "https://www.babelio.com")
	s = strings.TrimPrefix(s, "http://www.babelio.com")
	s = strings.TrimPrefix(s, "www.babelio.com")
	if !strings.HasPrefix(s, "/") {
		s = "/livres/" + s
	}
	// Drop any trailing segment after /livres/SLUG/ID (e.g. /critiques)
	if m := bookPathRE.FindString(s); m != "" {
		return m, nil
	}
	return "", errors.New("cannot extract book id from input (expected URL, path, or slug/id)")
}

func ParseBook(doc *goquery.Document, url string) BookDetail {
	b := BookDetail{URL: url}
	if m := bookPathRE.FindStringSubmatch(url); m != nil {
		b.BookID = m[1]
	}
	b.Title = strings.TrimSpace(doc.Find(`h1[itemprop="name"], h1`).First().Text())
	authorLink := doc.Find(`a[href*="/auteur/"]`).First()
	b.Author = strings.TrimSpace(authorLink.Text())
	if href, ok := authorLink.Attr("href"); ok {
		if m := auteurRE.FindStringSubmatch(href); m != nil {
			b.AuthorID = m[1]
		}
	}
	b.Synopsis = trimSynopsis(cleanText(doc.Find(`[itemprop="description"], #d_bio, .livre_resume`).First().Text()))
	if rv := strings.TrimSpace(doc.Find(`.grosse_note, [itemprop="ratingValue"]`).First().Text()); rv != "" {
		rv = strings.ReplaceAll(rv, ",", ".")
		if f, err := strconv.ParseFloat(rv, 64); err == nil {
			b.AvgRating = f
		}
	}
	if n := strings.TrimSpace(doc.Find(`[itemprop="ratingCount"]`).First().Text()); n != "" {
		if v, err := strconv.Atoi(stripNonDigits(n)); err == nil {
			b.NbRatings = v
		}
	}
	b.Publisher = strings.TrimSpace(doc.Find(`a[href*="editeur"]`).First().Text())
	bodyText := doc.Find("body").Text()
	if m := nbPagesRE.FindStringSubmatch(bodyText); m != nil {
		if v, err := strconv.Atoi(m[1]); err == nil {
			b.Pages = v
		}
	}
	// "Critiques (NN)" link in the book's main nav — reliable per-book count.
	navCritiques := doc.Find(`a.menu_link[href*="/critiques"]`).First().Text()
	if m := regexp.MustCompile(`\((\d+)\)`).FindStringSubmatch(navCritiques); m != nil {
		if v, err := strconv.Atoi(m[1]); err == nil {
			b.NbReviews = v
		}
	}
	if m := datePubRE.FindStringSubmatch(bodyText); m != nil {
		b.DatePub = strings.TrimSpace(m[1])
		if i := strings.Index(b.DatePub, "\t"); i > 0 {
			b.DatePub = b.DatePub[:i]
		}
	}
	doc.Find(`a[href*="/categorie/"], a[href*="/rayon/"], a[href*="/genre/"]`).Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" {
			b.Genres = append(b.Genres, t)
		}
	})
	return b
}

// trimSynopsis drops the "Voir plus" footer and the "Contributeurs : …" line
// that Babelio appends to the description block.
func trimSynopsis(s string) string {
	for _, cut := range []string{">Voir plus", "Voir plus", "Contributeurs :"} {
		if i := strings.Index(s, cut); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r", "")
	// collapse whitespace runs
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func stripNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
