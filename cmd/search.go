package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/babelio"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Babelio for books",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		if strings.TrimSpace(query) == "" {
			return errors.New("empty query")
		}
		c, err := client.New()
		if err != nil {
			return err
		}
		form := url.Values{}
		form.Set("Recherche", query)
		form.Set("recherche", "OK")
		doc, err := c.PostForm("/recherche.php", form)
		if err != nil {
			return err
		}
		results := babelio.ParseSearchResults(doc)
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(map[string]any{
			"query":   query,
			"count":   len(results),
			"results": results,
		}, func(w io.Writer) error {
			for _, r := range results {
				star := ""
				if r.Rating > 0 {
					star = fmt.Sprintf("%.1f ", r.Rating)
				}
				if _, err := fmt.Fprintf(w, "%s%s — %s  (%s)\n", star, r.Title, r.Author, r.URL); err != nil {
					return err
				}
			}
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
