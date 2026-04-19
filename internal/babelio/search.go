package babelio

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SearchResult struct {
	BookID   string  `json:"book_id"`
	Title    string  `json:"title"`
	Author   string  `json:"author,omitempty"`
	AuthorID string  `json:"author_id,omitempty"`
	URL      string  `json:"url"`
	Rating   float64 `json:"avg_rating,omitempty"`
}

// ParseSearchResults extracts book cards from /recherche.php results.
// Babelio uses two different layouts depending on the query:
//   - `div.cr_carte`: rich card (title match or mixed query)
//   - `li.item` > `div.fiche_livreH`: mosaic tile (exact author match)
// Both are parsed; results are de-duplicated by book ID.
func ParseSearchResults(doc *goquery.Document) []SearchResult {
	var results []SearchResult
	seen := map[string]bool{}
	add := func(card *goquery.Selection) {
		bookLink := card.Find(`a[href*="/livres/"]`).First()
		href, ok := bookLink.Attr("href")
		if !ok {
			return
		}
		m := livresRE.FindStringSubmatch(href)
		if m == nil {
			return
		}
		id := m[1]
		if seen[id] {
			return
		}
		seen[id] = true

		r := SearchResult{
			BookID: id,
			URL:    absoluteURL(href),
			Title:  strings.TrimSpace(card.Find(".titre1").First().Text()),
		}
		authorLink := card.Find(`a[href*="/auteur/"]`).First()
		r.Author = strings.TrimSpace(authorLink.Text())
		if ah, ok := authorLink.Attr("href"); ok {
			if am := auteurRE.FindStringSubmatch(ah); am != nil {
				r.AuthorID = am[1]
			}
		}
		if rv, ok := card.Find("div.rateit").First().Attr("data-rateit-value"); ok {
			if f, err := strconv.ParseFloat(rv, 64); err == nil {
				r.Rating = f
			}
		}
		results = append(results, r)
	}
	doc.Find("div.cr_carte").Each(func(_ int, c *goquery.Selection) { add(c) })
	doc.Find("li.item").Each(func(_ int, li *goquery.Selection) {
		if li.Find("div.fiche_livreH").Length() > 0 {
			add(li)
		}
	})
	return results
}
