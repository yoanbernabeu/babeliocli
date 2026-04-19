package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/babeliocli/internal/client"
	"github.com/yoanbernabeu/babeliocli/internal/output"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage the locally stored Babelio session",
}

var (
	imPhpSessID  string
	imBbac       string
	imBbacml     string
	imIDUser     string
	imUsername   string
	imCookieFile string
)

var sessionImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import cookies from a browser session (useful for Google/Facebook SSO)",
	Long: `Import an existing Babelio session into the CLI instead of logging in with
a username and password. Useful when your account is linked to Google or
Facebook SSO.

Two ways to provide cookies:

  1. Flag-by-flag, copy-pasted from the browser DevTools (Application > Cookies):
       --phpsessid, --bbac, --id-user, [--bbacml], [--username]

  2. A JSON file exported by a cookie-export extension (e.g. "EditThisCookie"):
       --cookie-file cookies.json
     The JSON must be an array of objects with at least "name" and "value"
     fields. "domain", "path", "secure", "httpOnly", "expirationDate" are used
     if present.

The session is written to $XDG_CONFIG_HOME/babeliocli/session.json (mode 0600)
and verified against /mabibliotheque.php.`,
	RunE: runSessionImport,
}

func init() {
	sessionImportCmd.Flags().StringVar(&imPhpSessID, "phpsessid", "", "PHPSESSID cookie value")
	sessionImportCmd.Flags().StringVar(&imBbac, "bbac", "", "bbac cookie value")
	sessionImportCmd.Flags().StringVar(&imBbacml, "bbacml", "", "bbacml cookie value (optional)")
	sessionImportCmd.Flags().StringVar(&imIDUser, "id-user", "", "id_user cookie value")
	sessionImportCmd.Flags().StringVar(&imUsername, "username", "", "display name for the session (optional)")
	sessionImportCmd.Flags().StringVar(&imCookieFile, "cookie-file", "", "path to a JSON cookie export (overrides individual flags)")
	sessionCmd.AddCommand(sessionImportCmd)
	rootCmd.AddCommand(sessionCmd)
}

func runSessionImport(cmd *cobra.Command, _ []string) error {
	var cookies []*http.Cookie
	var err error
	if imCookieFile != "" {
		cookies, err = cookiesFromFile(imCookieFile)
		if err != nil {
			return fmt.Errorf("read cookie file: %w", err)
		}
	} else {
		cookies, err = cookiesFromFlags()
		if err != nil {
			return err
		}
	}
	if len(cookies) == 0 {
		return errors.New("no cookies to import")
	}
	c, err := client.ImportCookies(cookies, imUsername)
	if err != nil {
		return err
	}
	p := output.Printer{Format: outputFormat, Out: cmd.OutOrStdout()}
	return p.Emit(map[string]any{
		"status":   "ok",
		"username": c.Username(),
		"imported": len(cookies),
	}, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, "Session imported (%d cookies) for %q\n", len(cookies), c.Username())
		return err
	})
}

func cookiesFromFlags() ([]*http.Cookie, error) {
	if imPhpSessID == "" || imBbac == "" || imIDUser == "" {
		return nil, errors.New("--phpsessid, --bbac and --id-user are required (or use --cookie-file)")
	}
	base := []*http.Cookie{
		{Name: "PHPSESSID", Value: imPhpSessID, Domain: ".babelio.com", Path: "/"},
		{Name: "bbac", Value: imBbac, Domain: ".babelio.com", Path: "/"},
		{Name: "id_user", Value: imIDUser, Domain: ".babelio.com", Path: "/"},
	}
	if imBbacml != "" {
		base = append(base, &http.Cookie{Name: "bbacml", Value: imBbacml, Domain: ".babelio.com", Path: "/"})
	}
	return base, nil
}

type exportedCookie struct {
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Domain         string  `json:"domain"`
	Path           string  `json:"path"`
	Secure         bool    `json:"secure"`
	HTTPOnly       bool    `json:"httpOnly"`
	ExpirationDate float64 `json:"expirationDate"`
}

func cookiesFromFile(path string) ([]*http.Cookie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []exportedCookie
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("expected a JSON array of cookie objects: %w", err)
	}
	out := make([]*http.Cookie, 0, len(raw))
	for _, r := range raw {
		domain := r.Domain
		if domain != "" && !strings.Contains(domain, "babelio.com") {
			continue
		}
		if domain == "" {
			domain = ".babelio.com"
		}
		path := r.Path
		if path == "" {
			path = "/"
		}
		out = append(out, &http.Cookie{
			Name:     r.Name,
			Value:    r.Value,
			Domain:   domain,
			Path:     path,
			Secure:   r.Secure,
			HttpOnly: r.HTTPOnly,
		})
	}
	return out, nil
}
