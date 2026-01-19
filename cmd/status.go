package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
	"github.com/user/tcrs/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show login status",
	Long:  `Show the current login status and session information.`,
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	userID := findLoggedInUser()
	if userID == "" {
		if IsJSON() {
			result := map[string]interface{}{
				"logged_in": false,
				"message":   "Not logged in",
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println("Not logged in")
		}
		return
	}

	c, err := client.NewClient(userID, cfg)
	if err != nil {
		if IsJSON() {
			result := map[string]interface{}{
				"logged_in": false,
				"error":     err.Error(),
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	sessionInfo, err := c.GetSessionInfo()
	if err != nil {
		if IsJSON() {
			result := map[string]interface{}{
				"logged_in": false,
				"user_id":   userID,
				"error":     err.Error(),
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Session info error: %v\n", err)
		}
		return
	}

	// Check if session is still valid (not expired)
	sessionAge := time.Since(sessionInfo.CreatedAt)
	isExpired := sessionAge > time.Duration(config.SessionTimeout)*time.Hour
	expiresIn := time.Duration(config.SessionTimeout)*time.Hour - sessionAge

	if IsJSON() {
		result := map[string]interface{}{
			"logged_in":    !isExpired,
			"user_id":      userID,
			"created_at":   sessionInfo.CreatedAt.Format(time.RFC3339),
			"session_age":  sessionAge.String(),
			"expires_in":   expiresIn.String(),
			"is_expired":   isExpired,
			"cookie_count": sessionInfo.CookieCount,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		if isExpired {
			fmt.Printf("Session expired for user: %s\n", userID)
			fmt.Printf("  Session created: %s\n", sessionInfo.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("  Please login again")
		} else {
			fmt.Printf("Logged in as: %s\n", userID)
			fmt.Printf("  Session created: %s\n", sessionInfo.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Session age: %s\n", formatDuration(sessionAge))
			fmt.Printf("  Expires in: %s\n", formatDuration(expiresIn))
		}
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
