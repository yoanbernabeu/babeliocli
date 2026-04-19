package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const babelioURL = "https://www.babelio.com"

type storedCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"httpOnly"`
}

type Session struct {
	Cookies   []storedCookie `json:"cookies"`
	IDUser    string         `json:"id_user,omitempty"`
	Username  string         `json:"username,omitempty"`
	SavedAt   time.Time      `json:"saved_at"`
}

func sessionDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "babeliocli"), nil
}

func sessionPath() (string, error) {
	dir, err := sessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func SaveSession(jar http.CookieJar, username string) error {
	u, _ := url.Parse(babelioURL)
	cookies := jar.Cookies(u)
	stored := make([]storedCookie, 0, len(cookies))
	idUser := ""
	for _, c := range cookies {
		stored = append(stored, storedCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
		if c.Name == "id_user" {
			idUser = c.Value
		}
	}
	sess := Session{
		Cookies:  stored,
		IDUser:   idUser,
		Username: username,
		SavedAt:  time.Now(),
	}
	dir, err := sessionDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path, err := sessionPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(&sess)
}

func LoadSession() (*Session, http.CookieJar, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, ErrNoSession
		}
		return nil, nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, nil, fmt.Errorf("invalid session file: %w", err)
	}
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(babelioURL)
	cookies := make([]*http.Cookie, 0, len(sess.Cookies))
	for _, c := range sess.Cookies {
		cookies = append(cookies, &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
	}
	jar.SetCookies(u, cookies)
	return &sess, jar, nil
}

func DeleteSession() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

var ErrNoSession = errors.New("no saved session (run `babeliocli login` first)")
