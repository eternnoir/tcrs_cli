// Package client provides HTTP client functionality for TCRS.
package client

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/user/tcrs/internal/config"
)

// SessionInfo holds session metadata.
type SessionInfo struct {
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	CookieCount int       `json:"cookie_count"`
}

// CookieData represents a serializable cookie.
type CookieData struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  int64  `json:"expires"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
}

// SessionManager handles cookie persistence and session validation.
type SessionManager struct {
	userID  string
	cfg     *config.Config
	jar     http.CookieJar
	baseURL *url.URL
}

// NewSessionManager creates a new session manager.
func NewSessionManager(userID string, cfg *config.Config) (*SessionManager, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	sm := &SessionManager{
		userID:  userID,
		cfg:     cfg,
		jar:     jar,
		baseURL: baseURL,
	}

	// Try to load existing cookies
	_ = sm.loadCookies()

	return sm, nil
}

// CookieJar returns the HTTP cookie jar.
func (sm *SessionManager) CookieJar() http.CookieJar {
	return sm.jar
}

// loadCookies loads cookies from disk if they exist and are valid.
func (sm *SessionManager) loadCookies() error {
	// Check if session info file exists and is valid
	sessionFile := sm.cfg.SessionFile(sm.userID)
	sessionData, err := os.ReadFile(sessionFile)
	if err != nil {
		return err
	}

	var sessionInfo SessionInfo
	if err := json.Unmarshal(sessionData, &sessionInfo); err != nil {
		return err
	}

	// Check if session is expired (12 hours)
	if time.Since(sessionInfo.CreatedAt) > time.Duration(config.SessionTimeout)*time.Hour {
		return ErrSessionExpired
	}

	// Load cookies
	cookieFile := sm.cfg.CookieFile(sm.userID)
	cookieData, err := os.ReadFile(cookieFile)
	if err != nil {
		return err
	}

	var cookies []CookieData
	if err := json.Unmarshal(cookieData, &cookies); err != nil {
		return err
	}

	// Convert to http.Cookie and add to jar
	httpCookies := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		httpCookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
		if c.Expires > 0 {
			httpCookie.Expires = time.Unix(c.Expires, 0)
		}
		httpCookies = append(httpCookies, httpCookie)
	}

	sm.jar.SetCookies(sm.baseURL, httpCookies)
	return nil
}

// SaveCookies saves current cookies to disk.
func (sm *SessionManager) SaveCookies() error {
	if err := sm.cfg.EnsureCacheDir(); err != nil {
		return err
	}

	cookies := sm.jar.Cookies(sm.baseURL)
	if len(cookies) == 0 {
		return ErrNoCookies
	}

	// Check for session cookie
	hasSession := false
	for _, c := range cookies {
		if isSessionCookie(c.Name) {
			hasSession = true
			break
		}
	}
	if !hasSession {
		return ErrNoSessionCookie
	}

	// Convert to serializable format
	cookieData := make([]CookieData, 0, len(cookies))
	for _, c := range cookies {
		cd := CookieData{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
		if !c.Expires.IsZero() {
			cd.Expires = c.Expires.Unix()
		}
		cookieData = append(cookieData, cd)
	}

	// Save cookies
	cookieJSON, err := json.MarshalIndent(cookieData, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(sm.cfg.CookieFile(sm.userID), cookieJSON, 0600); err != nil {
		return err
	}

	// Save session info
	sessionInfo := SessionInfo{
		UserID:      sm.userID,
		CreatedAt:   time.Now(),
		CookieCount: len(cookies),
	}
	sessionJSON, err := json.MarshalIndent(sessionInfo, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sm.cfg.SessionFile(sm.userID), sessionJSON, 0600)
}

// ClearCookies removes all saved cookies.
func (sm *SessionManager) ClearCookies() error {
	// Create a new empty jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	sm.jar = jar

	// Remove cookie file
	cookieFile := sm.cfg.CookieFile(sm.userID)
	if err := os.Remove(cookieFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove session file
	sessionFile := sm.cfg.SessionFile(sm.userID)
	if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// HasValidSession checks if there's a valid session.
func (sm *SessionManager) HasValidSession() bool {
	cookies := sm.jar.Cookies(sm.baseURL)
	for _, c := range cookies {
		if isSessionCookie(c.Name) {
			return true
		}
	}
	return false
}

// GetSessionInfo returns the current session info if exists.
func (sm *SessionManager) GetSessionInfo() (*SessionInfo, error) {
	sessionFile := sm.cfg.SessionFile(sm.userID)
	sessionData, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, err
	}

	var sessionInfo SessionInfo
	if err := json.Unmarshal(sessionData, &sessionInfo); err != nil {
		return nil, err
	}

	return &sessionInfo, nil
}

// isSessionCookie checks if a cookie name is a session cookie.
func isSessionCookie(name string) bool {
	sessionCookieNames := []string{
		"JSESSIONID", "session", "sessionid", "sid",
		"_session_id", "ASP.NET_SessionId", "PHPSESSID",
	}
	nameLower := strings.ToLower(name)
	for _, scn := range sessionCookieNames {
		if nameLower == strings.ToLower(scn) {
			return true
		}
	}
	return false
}
