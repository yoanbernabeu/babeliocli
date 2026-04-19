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

var bookCmd = &cobra.Command{
	Use:   "book <url-or-path>",
	Short: "Show details of a book by URL, path, or slug/id",
	Long: `Accepts any of:
  - https://www.babelio.com/livres/Simmons-Hyperion-1/5603
  - /livres/Simmons-Hyperion-1/5603
  - Simmons-Hyperion-1/5603`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := babelio.ResolveBookPath(args[0])
		if err != nil {
			return err
		}
		c, err := client.New()
		if err != nil {
			return err
		}
		doc, err := c.Get(path)
		if err != nil {
			return err
		}
		b := babelio.ParseBook(doc, "https://www.babelio.com"+path)
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(b, func(w io.Writer) error {
			fmt.Fprintf(w, "%s\n", b.Title)
			if b.Author != "" {
				fmt.Fprintf(w, "by %s\n", b.Author)
			}
			if b.AvgRating > 0 {
				fmt.Fprintf(w, "%.2f/5 (%d notes, %d critiques)\n", b.AvgRating, b.NbRatings, b.NbReviews)
			}
			if b.Publisher != "" || b.Pages > 0 {
				var parts []string
				if b.Publisher != "" {
					parts = append(parts, b.Publisher)
				}
				if b.Pages > 0 {
					parts = append(parts, fmt.Sprintf("%d pages", b.Pages))
				}
				fmt.Fprintf(w, "%s\n", strings.Join(parts, " — "))
			}
			if b.Synopsis != "" {
				fmt.Fprintf(w, "\n%s\n", b.Synopsis)
			}
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(bookCmd)
}
