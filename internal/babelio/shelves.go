package babelio

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Shelf struct {
	Key   string `json:"key"`   // slug: lus, a-lire, etc.
	Label string `json:"label"` // raw label from the site
	Count int    `json:"count"`
	URL   string `json:"url"` // path with query string
}

var shelfKeyMap = map[string]string{
	"s1=0":         "en-cours",
	"s1=1":         "lus",
	"s1=2":         "a-lire",
	"s1=3":         "pense-bete",
	"s1=4":         "abandonnes",
	"s3=10":        "critiques",
	"s3=11":        "non-critiques",
	"action=toute": "all",
}

var countRE = regexp.MustCompile(`\((\d+)\)`)

// ParseShelves extracts the shelf list from /mabibliotheque.php. Only the
// standard shelves (lus, à lire, en cours, etc.) are returned; sort options
// (filtre=titre, filtre=auteur) and rating buckets (note=) are skipped.
func ParseShelves(doc *goquery.Document) []Shelf {
	var shelves []Shelf
	seen := map[string]bool{}
	doc.Find("select option").Each(func(_ int, s *goquery.Selection) {
		val, ok := s.Attr("value")
		if !ok || !strings.Contains(val, "mabibliotheque.php") {
			return
		}
		key := shelfKeyFromURL(val)
		if key == "" {
			return
		}
		if seen[key] {
			return
		}
		seen[key] = true
		label := strings.TrimSpace(s.Text())
		count := 0
		if m := countRE.FindStringSubmatch(label); m != nil {
			count, _ = strconv.Atoi(m[1])
		}
		labelClean := strings.TrimSpace(countRE.ReplaceAllString(label, ""))
		shelves = append(shelves, Shelf{
			Key:   key,
			Label: labelClean,
			Count: count,
			URL:   normalizePath(val),
		})
	})
	return shelves
}

func shelfKeyFromURL(val string) string {
	low := strings.ToLower(val)
	for needle, key := range shelfKeyMap {
		if strings.Contains(low, needle) {
			return key
		}
	}
	return ""
}

// ShelfURL returns the URL for a given shelf key (used by `books` command).
func ShelfURL(key string) (string, bool) {
	for needle, k := range shelfKeyMap {
		if k == key {
			if k == "all" {
				return "/mabibliotheque.php?action=toute", true
			}
			return "/mabibliotheque.php?" + needle, true
		}
	}
	return "", false
}

// KnownShelfKeys returns supported shelf keys in display order.
func KnownShelfKeys() []string {
	return []string{"all", "lus", "a-lire", "en-cours", "pense-bete", "abandonnes", "critiques", "non-critiques"}
}

func normalizePath(val string) string {
	// Site returns values like "mabibliotheque.php?&s1=1" — make it absolute path.
	v := val
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	v = strings.Replace(v, "?&", "?", 1)
	return v
}
