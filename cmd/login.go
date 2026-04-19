package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
	"golang.org/x/term"
)

var (
	loginUsername      string
	loginPassword      string
	loginPasswordStdin bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate against babelio.com and store the session locally",
	Long: `Log in to babelio.com. Credentials can be provided via flags,
environment variables (BABELIO_USERNAME, BABELIO_PASSWORD), --password-stdin,
or interactive prompt. The cookie jar is stored at
$XDG_CONFIG_HOME/babeliocli/session.json (mode 0600).`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "babelio username or email (env BABELIO_USERNAME)")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "babelio password (env BABELIO_PASSWORD). Prefer --password-stdin")
	loginCmd.Flags().BoolVar(&loginPasswordStdin, "password-stdin", false, "read password from stdin (recommended for scripts)")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	user := strings.TrimSpace(loginUsername)
	if user == "" {
		user = os.Getenv("BABELIO_USERNAME")
	}
	pass := loginPassword
	if pass == "" {
		pass = os.Getenv("BABELIO_PASSWORD")
	}
	if loginPasswordStdin {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read password from stdin: %w", err)
		}
		pass = strings.TrimRight(string(b), "\r\n")
	}

	interactive := term.IsTerminal(int(syscall.Stdin))
	if user == "" {
		if !interactive {
			return errors.New("missing username (use --username or BABELIO_USERNAME)")
		}
		fmt.Fprint(os.Stderr, "Babelio username/email: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		user = strings.TrimSpace(line)
	}
	if pass == "" {
		if !interactive {
			return errors.New("missing password (use --password, --password-stdin, or BABELIO_PASSWORD)")
		}
		fmt.Fprint(os.Stderr, "Babelio password: ")
		b, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		pass = string(b)
	}

	c, err := client.Login(user, pass)
	if err != nil {
		return err
	}

	result := map[string]any{
		"status":   "ok",
		"username": c.Username(),
	}
	p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
	return p.Emit(result, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "Logged in as %s\n", c.Username())
		return err
	})
}
