// Package cmd provides CLI commands for TCRS.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/config"
)

var (
	// Version is set at build time.
	Version = "dev"

	// Global flags
	cfgFile string
	verbose bool
	jsonOut bool

	// Global config
	cfg *config.Config
)

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "tcrs",
	Short: "TCRS CLI - Timecard Recording System",
	Long: `TCRS CLI is a command-line tool for interacting with the
Timecard Recording System (TCRS).

It provides commands for:
  - Authentication (login, logout, status)
  - Querying projects and activities
  - Viewing and saving weekly timecards`,
	Version: Version,
}

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")
}

// initConfig initializes the global configuration.
func initConfig() {
	cfg = config.DefaultConfig()
	cfg.Verbose = verbose
	cfg.JSON = jsonOut
}

// GetConfig returns the global configuration.
func GetConfig() *config.Config {
	return cfg
}

// IsVerbose returns true if verbose output is enabled.
func IsVerbose() bool {
	return verbose
}

// IsJSON returns true if JSON output is enabled.
func IsJSON() bool {
	return jsonOut
}
