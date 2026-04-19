package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Delete the locally stored session",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := client.DeleteSession(); err != nil {
			return err
		}
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(map[string]any{"status": "ok"}, func(w io.Writer) error {
			_, err := fmt.Fprintln(w, "Session deleted")
			return err
		})
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
