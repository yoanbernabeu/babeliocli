package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/babelio"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var (
	booksShelf    string
	booksLimit    int
	booksMaxPages int
)

var booksCmd = &cobra.Command{
	Use:   "books",
	Short: "List books in a given shelf",
	Long: fmt.Sprintf(`List books from the authenticated user's library.

Supported shelf keys: %s.

Pagination is followed automatically; use --limit to cap the number of books
returned, or --max-pages to cap page fetches.`, strings.Join(babelio.KnownShelfKeys(), ", ")),
	RunE: runBooks,
}

func init() {
	booksCmd.Flags().StringVarP(&booksShelf, "shelf", "s", "all", "shelf key: "+strings.Join(babelio.KnownShelfKeys(), "|"))
	booksCmd.Flags().IntVar(&booksLimit, "limit", 0, "stop after N books (0 = no limit)")
	booksCmd.Flags().IntVar(&booksMaxPages, "max-pages", 50, "maximum pages to fetch (safety cap)")
	rootCmd.AddCommand(booksCmd)
}

func runBooks(cmd *cobra.Command, _ []string) error {
	c, err := client.New()
	if err != nil {
		return err
	}
	var all []babelio.Book
	fetched := 0
	for pageNum := 1; pageNum <= booksMaxPages; pageNum++ {
		u, err := babelio.ShelfPageURL(booksShelf, pageNum)
		if err != nil {
			return err
		}
		doc, err := c.Get(u)
		if err != nil {
			return err
		}
		page := babelio.ParseBooks(doc)
		if len(page.Books) == 0 {
			break
		}
		for _, b := range page.Books {
			all = append(all, b)
			fetched++
			if booksLimit > 0 && fetched >= booksLimit {
				break
			}
		}
		if booksLimit > 0 && fetched >= booksLimit {
			break
		}
		if pageNum >= page.TotalPages {
			break
		}
	}

	p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
	return p.Emit(map[string]any{
		"shelf": booksShelf,
		"count": len(all),
		"books": all,
	}, func(w io.Writer) error {
		for _, b := range all {
			star := strings.Repeat("★", b.Rating) + strings.Repeat("·", 5-b.Rating)
			if _, err := fmt.Fprintf(w, "%s  %s — %s [%s]\n", star, b.Title, b.Author, b.Status); err != nil {
				return err
			}
		}
		return nil
	})
}
