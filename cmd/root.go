package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "babeliocli",
	Short: "Read-only CLI for Babelio (unofficial)",
	Long: `babeliocli is a read-only command-line client for babelio.com.

Babelio does not provide an official API, so this tool authenticates via the
regular web form and parses the HTML of the logged-in pages. Output is JSON
by default, so it can be piped into jq or consumed by coding agents.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "json", "output format: json|text")
}
