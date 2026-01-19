package client

import "errors"

var (
	// ErrSessionExpired indicates the session has expired.
	ErrSessionExpired = errors.New("session expired")
	// ErrNoCookies indicates no cookies are present.
	ErrNoCookies = errors.New("no cookies found")
	// ErrNoSessionCookie indicates no session cookie was found.
	ErrNoSessionCookie = errors.New("no session cookie found")
	// ErrNotLoggedIn indicates the user is not logged in.
	ErrNotLoggedIn = errors.New("not logged in")
	// ErrLoginFailed indicates login failed.
	ErrLoginFailed = errors.New("login failed")
	// ErrInvalidCredentials indicates invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
