package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

const userAgent = "babeliocli/0.1 (+https://github.com/yoanbernabeu/babeliocli)"

type Client struct {
	http     *http.Client
	jar      http.CookieJar
	username string
}

func newHTTPClient(jar http.CookieJar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
}

func New() (*Client, error) {
	sess, jar, err := LoadSession()
	if err != nil {
		return nil, err
	}
	return &Client{
		http:     newHTTPClient(jar),
		jar:      jar,
		username: sess.Username,
	}, nil
}

func newAnonymous() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		http: newHTTPClient(jar),
		jar:  jar,
	}, nil
}

func (c *Client) Username() string {
	return c.username
}

// ErrSessionExpired is returned when a request requiring authentication gets
// redirected to the login page. Callers should prompt the user to re-import.
var ErrSessionExpired = errors.New("session expired: run `babeliocli login` or `babeliocli session import` again")

// Get fetches the given path, checks for a redirect to the login page, and
// returns a goquery document. The response body is always closed before
// returning.
func (c *Client) Get(path string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, babelioURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if strings.Contains(resp.Request.URL.Path, "connection.php") && !strings.Contains(path, "connection.php") {
		return nil, ErrSessionExpired
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset decode: %w", err)
	}
	return goquery.NewDocumentFromReader(reader)
}

// PostForm submits a POST with application/x-www-form-urlencoded body and
// returns the parsed document (charset-aware).
func (c *Client) PostForm(path string, form url.Values) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodPost, babelioURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	req.Header.Set("Referer", babelioURL+path)
	req.Header.Set("Origin", babelioURL)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("POST %s: HTTP %d", path, resp.StatusCode)
	}
	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset decode: %w", err)
	}
	return goquery.NewDocumentFromReader(reader)
}

var errLoginFailed = errors.New("login failed: check username and password")

// ImportCookies builds a session from an arbitrary list of cookies (typically
// copied from the user's browser after a Google/Facebook SSO login). It saves
// the session to disk and returns a ready-to-use Client. The session is
// verified by requesting /mabibliotheque.php and ensuring we aren't redirected
// back to the login page.
func ImportCookies(cookies []*http.Cookie, username string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(babelioURL)
	jar.SetCookies(u, cookies)
	c := &Client{http: newHTTPClient(jar), jar: jar, username: username}
	if err := c.verifyAuthenticated(); err != nil {
		return nil, err
	}
	if username == "" {
		c.username = c.detectUsername()
	}
	if err := SaveSession(c.jar, c.username); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return c, nil
}

// verifyAuthenticated returns nil if the stored cookies give access to the
// logged-in library page. Babelio redirects unauthenticated requests to
// /connection.php, so we follow redirects and check the final URL.
func (c *Client) verifyAuthenticated() error {
	req, err := http.NewRequest(http.MethodGet, babelioURL+"/mabibliotheque.php", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	final := resp.Request.URL.Path
	if strings.Contains(final, "connection.php") {
		return errors.New("imported cookies are not authenticated (redirected to login)")
	}
	return nil
}

// detectUsername tries to pull the displayed username from /monprofil.php.
// Best-effort: returns "" if parsing fails.
func (c *Client) detectUsername() string {
	doc, err := c.Get("/monprofil.php")
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(doc.Find(".pseudo, .pseudo_profil, h1.pseudo").First().Text())
	return name
}

func Login(username, password string) (*Client, error) {
	c, err := newAnonymous()
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("Login", username)
	form.Set("Password", password)
	form.Set("sub_btn", "connexion")
	form.Set("ref", "")

	req, err := http.NewRequest(http.MethodPost, babelioURL+"/connection.php?r=1", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", babelioURL+"/connection.php")
	req.Header.Set("Origin", babelioURL)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	u, _ := url.Parse(babelioURL)
	var hasIDUser bool
	for _, ck := range c.jar.Cookies(u) {
		if ck.Name == "id_user" && ck.Value != "" && ck.Value != "0" {
			hasIDUser = true
		}
	}
	if !hasIDUser {
		return nil, errLoginFailed
	}
	c.username = username
	if err := SaveSession(c.jar, username); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}
	return c, nil
}
