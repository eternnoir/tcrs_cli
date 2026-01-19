package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from TCRS",
	Long:  `Logout from TCRS and clear saved session cookies.`,
	Run:   runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) {
	// Find the user from saved session files
	userID := findLoggedInUser()
	if userID == "" {
		if IsJSON() {
			result := map[string]interface{}{
				"success": true,
				"message": "No active session found",
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println("No active session found")
		}
		return
	}

	c, err := client.NewClient(userID, cfg)
	if err != nil {
		printError("Failed to create client", err)
		os.Exit(1)
	}

	if IsVerbose() {
		fmt.Printf("Logging out %s...\n", userID)
	}

	err = c.Logout()
	if err != nil {
		printError("Logout failed", err)
		os.Exit(1)
	}

	if IsJSON() {
		result := map[string]interface{}{
			"success": true,
			"user_id": userID,
			"message": "Logout successful",
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Successfully logged out %s\n", userID)
	}
}

// findLoggedInUser finds the user ID from saved session files.
func findLoggedInUser() string {
	cacheDir := cfg.CacheDir

	// Look for .session files
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".session") {
			// Extract user ID from filename
			userID := strings.TrimSuffix(entry.Name(), ".session")
			// Verify cookie file exists
			cookieFile := filepath.Join(cacheDir, userID+".cookies")
			if _, err := os.Stat(cookieFile); err == nil {
				return userID
			}
		}
	}

	return ""
}
