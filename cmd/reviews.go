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
	reviewsLimit    int
	reviewsMaxPages int
)

var reviewsCmd = &cobra.Command{
	Use:   "reviews <url-or-path>",
	Short: "List critiques (reviews) for a given book",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := babelio.ResolveBookPath(args[0])
		if err != nil {
			return err
		}
		c, err := client.New()
		if err != nil {
			return err
		}
		base := path + "/critiques"
		var all []babelio.Review
		fetched := 0
		for pageNum := 1; pageNum <= reviewsMaxPages; pageNum++ {
			u := base
			if pageNum > 1 {
				u = fmt.Sprintf("%s?pageN=%d", base, pageNum)
			}
			doc, err := c.Get(u)
			if err != nil {
				return err
			}
			rp := babelio.ParseReviews(doc)
			if len(rp.Reviews) == 0 {
				break
			}
			for _, r := range rp.Reviews {
				all = append(all, r)
				fetched++
				if reviewsLimit > 0 && fetched >= reviewsLimit {
					break
				}
			}
			if reviewsLimit > 0 && fetched >= reviewsLimit {
				break
			}
			if pageNum >= rp.TotalPages {
				break
			}
		}
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(map[string]any{
			"book_url": "https://www.babelio.com" + path,
			"count":    len(all),
			"reviews":  all,
		}, func(w io.Writer) error {
			for _, r := range all {
				star := strings.Repeat("★", r.Rating) + strings.Repeat("·", 5-r.Rating)
				fmt.Fprintf(w, "\n%s  %s  (%s)\n", star, r.Author, r.Date)
				fmt.Fprintln(w, r.Body)
			}
			return nil
		})
	},
}

func init() {
	reviewsCmd.Flags().IntVar(&reviewsLimit, "limit", 20, "stop after N reviews (0 = no limit)")
	reviewsCmd.Flags().IntVar(&reviewsMaxPages, "max-pages", 20, "maximum pages to fetch")
	rootCmd.AddCommand(reviewsCmd)
}
