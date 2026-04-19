package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/babelio"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var shelvesCmd = &cobra.Command{
	Use:   "shelves",
	Short: "List the authenticated user's shelves (étagères)",
	Long: `List the built-in Babelio shelves (Lus, À lire, En cours, etc.) and their
counts as extracted from the /mabibliotheque.php page.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}
		doc, err := c.Get("/mabibliotheque.php")
		if err != nil {
			return err
		}
		shelves := babelio.ParseShelves(doc)
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(map[string]any{"shelves": shelves}, func(w io.Writer) error {
			for _, s := range shelves {
				if _, err := fmt.Fprintf(w, "%-15s %-18s (%d)\n", s.Key, s.Label, s.Count); err != nil {
					return err
				}
			}
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(shelvesCmd)
}
