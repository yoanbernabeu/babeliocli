package babelio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Book struct {
	IDBiblio  string   `json:"id_biblio"`
	BookID    string   `json:"book_id,omitempty"`
	Title     string   `json:"title"`
	Author    string   `json:"author,omitempty"`
	AuthorID  string   `json:"author_id,omitempty"`
	BookURL   string   `json:"book_url,omitempty"`
	Status    string   `json:"status,omitempty"`     // À lire, Lu, etc.
	Rating    int      `json:"rating,omitempty"`     // 0-5, 0 = not rated
	Readers   int      `json:"readers,omitempty"`    // nb lecteurs
	Tags      []string `json:"tags,omitempty"`
	ReadStart string   `json:"read_start,omitempty"` // YYYY-MM-DD
	ReadEnd   string   `json:"read_end,omitempty"`   // YYYY-MM-DD
}

type BookPage struct {
	Books       []Book `json:"books"`
	Page        int    `json:"page"`
	TotalPages  int    `json:"total_pages,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
}

var (
	idBiblioRE = regexp.MustCompile(`id_biblio=(\d+)`)
	livresRE   = regexp.MustCompile(`/livres/[^/]+/(\d+)`)
	auteurRE   = regexp.MustCompile(`/auteur/[^/]+/(\d+)`)
	pageNRE    = regexp.MustCompile(`pageN=(\d+)`)
)

// ParseBooks extracts books from a /mabibliotheque.php result page.
func ParseBooks(doc *goquery.Document) BookPage {
	var page BookPage
	seen := map[string]bool{}

	doc.Find(`td.supprimer[id^="ligne_"]`).Each(func(_ int, td *goquery.Selection) {
		row := td.Closest("tr")
		if row.Length() == 0 {
			return
		}
		var b Book
		// id_biblio lives in the "supprimer" action link, not in the anchor id.
		if href, ok := row.Find(`a[href*="id_biblio="]`).First().Attr("href"); ok {
			if m := idBiblioRE.FindStringSubmatch(href); m != nil {
				b.IDBiblio = m[1]
			}
		}
		if b.IDBiblio == "" || seen[b.IDBiblio] {
			return
		}
		seen[b.IDBiblio] = true

		titleCell := row.Find("td.titre_livre").First()
		titleLink := titleCell.Find(`a[href*="/livres/"]`).First()
		b.Title = strings.TrimSpace(titleLink.Text())
		if href, ok := titleLink.Attr("href"); ok {
			b.BookURL = absoluteURL(href)
			if m := livresRE.FindStringSubmatch(href); m != nil {
				b.BookID = m[1]
			}
		}

		authorLink := row.Find(`td.auteur a[href*="/auteur/"]`).First()
		b.Author = strings.TrimSpace(authorLink.Text())
		if href, ok := authorLink.Attr("href"); ok {
			if m := auteurRE.FindStringSubmatch(href); m != nil {
				b.AuthorID = m[1]
			}
		}

		// Rating: div.rateit with data-rateit-value
		if rv, ok := row.Find("div.rateit").First().Attr("data-rateit-value"); ok {
			if f, err := strconv.ParseFloat(rv, 64); err == nil {
				b.Rating = int(f + 0.5) // round to nearest int (0..5)
			}
		}

		statusCell := row.Find("td.statut").First()
		if statusCell.Length() > 0 {
			txt := firstNonEmptyLine(statusCell.Text())
			b.Status = txt
			if v, ok := statusCell.Find(`input[class*="datepicker_deb"]`).First().Attr("value"); ok {
				b.ReadStart = parseFRDate(v)
			}
			if v, ok := statusCell.Find(`input[class*="datepicker_fin"]`).First().Attr("value"); ok {
				b.ReadEnd = parseFRDate(v)
			}
		}

		readersCell := row.Find("td.lecteurs").First()
		if readersCell.Length() > 0 {
			if n, err := strconv.Atoi(strings.TrimSpace(readersCell.Text())); err == nil {
				b.Readers = n
			}
		}

		// Tags/étiquettes in td.check / td.etiquette (if any rendered text)
		row.Find("td.etiquette a, td.check a").Each(func(_ int, a *goquery.Selection) {
			txt := strings.TrimSpace(a.Text())
			if txt == "" || txt == "+" {
				return
			}
			b.Tags = append(b.Tags, txt)
		})

		page.Books = append(page.Books, b)
	})

	// Pagination: look at links with pageN= in href
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

// parseFRDate converts a "DD/MM/YYYY" string to "YYYY-MM-DD". Returns "" on
// empty input or parse failure.
func parseFRDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return ""
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 || len(parts[2]) != 4 {
		return ""
	}
	return parts[2] + "-" + parts[1] + "-" + parts[0]
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			return t
		}
	}
	return ""
}

func absoluteURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return "https://www.babelio.com" + href
}

// ShelfPageURL builds the URL to request a specific page of a shelf.
//
// Babelio stores the active shelf filter server-side in the PHP session, so
// only the first request carries the shelf parameter (s1=, s3=, action=…).
// Subsequent pages must be requested with just pageN= to keep the filter.
func ShelfPageURL(shelfKey string, page int) (string, error) {
	if page < 1 {
		page = 1
	}
	if page == 1 {
		base, ok := ShelfURL(shelfKey)
		if !ok {
			return "", fmt.Errorf("unknown shelf %q", shelfKey)
		}
		return base, nil
	}
	return fmt.Sprintf("/mabibliotheque.php?pageN=%d", page), nil
}
