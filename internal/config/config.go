// Package config provides configuration management for TCRS CLI.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultBaseURL is the default TCRS API base URL.
	// Set TCRS_BASE_URL environment variable to override.
	DefaultBaseURL = ""
	// SessionTimeout is the session timeout in hours.
	SessionTimeout = 12
)

// Config holds the application configuration.
type Config struct {
	BaseURL  string
	CacheDir string
	Verbose  bool
	JSON     bool
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:  getEnvOrDefault("TCRS_BASE_URL", DefaultBaseURL),
		CacheDir: getEnvOrDefault("TCRS_CACHE_DIR", defaultCacheDir()),
		Verbose:  false,
		JSON:     false,
	}
}

// defaultCacheDir returns the default cache directory path.
func defaultCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".tcrs"
	}
	return filepath.Join(homeDir, ".tcrs")
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// EnsureCacheDir creates the cache directory if it doesn't exist.
func (c *Config) EnsureCacheDir() error {
	return os.MkdirAll(c.CacheDir, 0700)
}

// CookieFile returns the path to the cookie file for a user.
func (c *Config) CookieFile(userID string) string {
	return filepath.Join(c.CacheDir, userID+".cookies")
}

// SessionFile returns the path to the session info file for a user.
func (c *Config) SessionFile(userID string) string {
	return filepath.Join(c.CacheDir, userID+".session")
}

// ValidateBaseURL checks if the base URL is configured.
func (c *Config) ValidateBaseURL() error {
	if c.BaseURL == "" {
		return ErrBaseURLNotSet
	}
	return nil
}

// ErrBaseURLNotSet indicates TCRS_BASE_URL is not set.
var ErrBaseURLNotSet = fmt.Errorf("TCRS_BASE_URL environment variable is not set")

