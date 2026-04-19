package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently authenticated babelio account",
	RunE: func(cmd *cobra.Command, _ []string) error {
		sess, _, err := client.LoadSession()
		if err != nil {
			return err
		}
		out := map[string]any{
			"username":     sess.Username,
			"id_user":      sess.IDUser,
			"session_file": "~/.config/babeliocli/session.json",
			"saved_at":     sess.SavedAt,
		}
		p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
		return p.Emit(out, func(w io.Writer) error {
			_, err := fmt.Fprintf(w, "Logged in as %s (id=%s)\n", sess.Username, sess.IDUser)
			return err
		})
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
