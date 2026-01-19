package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
)

var loginCmd = &cobra.Command{
	Use:   "login [user_id] [password]",
	Short: "Login to TCRS",
	Long: `Authenticate with TCRS using your user ID and password.

Credentials can be provided as arguments or via environment variables:
  TCRS_USER     - User ID
  TCRS_PASSWORD - Password

Arguments take precedence over environment variables.`,
	Args: cobra.MaximumNArgs(2),
	Run:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) {
	// Get credentials from args or environment
	userID := os.Getenv("TCRS_USER")
	password := os.Getenv("TCRS_PASSWORD")

	if len(args) >= 1 {
		userID = args[0]
	}
	if len(args) >= 2 {
		password = args[1]
	}

	// Validate credentials
	if userID == "" {
		printError("Missing user ID", fmt.Errorf("provide as argument or set TCRS_USER"))
		os.Exit(1)
	}
	if password == "" {
		printError("Missing password", fmt.Errorf("provide as argument or set TCRS_PASSWORD"))
		os.Exit(1)
	}

	c, err := client.NewClient(userID, cfg)
	if err != nil {
		printError("Failed to create client", err)
		os.Exit(1)
	}

	if IsVerbose() {
		fmt.Printf("Logging in as %s...\n", userID)
	}

	err = c.Login(password)
	if err != nil {
		printError("Login failed", err)
		os.Exit(1)
	}

	if IsJSON() {
		result := map[string]interface{}{
			"success": true,
			"user_id": userID,
			"message": "Login successful",
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Successfully logged in as %s\n", userID)
	}
}

func printError(msg string, err error) {
	if IsJSON() {
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"message": msg,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	}
}
