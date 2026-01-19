package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
)

var (
	saveDate string
	saveFile string
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save timecard entries",
	Long: `Save timecard entries for a week from a JSON file or stdin.

The JSON format should be:
{
  "entries": [
    {
      "project_id": "12345",
      "activity_id": "5",
      "progress": 0,
      "days": [
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 0, "note": "", "progress": 0},
        {"hours": 0, "note": "", "progress": 0}
      ]
    }
  ]
}

Use "-" as the file argument to read from stdin.`,
	Run: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
	saveCmd.Flags().StringVar(&saveDate, "date", "", "week start date in YYYY-MM-DD format (default: this week's Monday)")
	saveCmd.Flags().StringVarP(&saveFile, "file", "f", "", "JSON file with entries (use '-' for stdin)")
	saveCmd.MarkFlagRequired("file")
}

// SaveInput represents the input JSON structure.
type SaveInput struct {
	Entries []client.SaveEntry `json:"entries"`
}

func runSave(cmd *cobra.Command, args []string) {
	userID := findLoggedInUser()
	if userID == "" {
		printError("Not logged in", fmt.Errorf("please login first with: tcrs login <user> <pass>"))
		os.Exit(1)
	}

	c, err := client.NewClient(userID, cfg)
	if err != nil {
		printError("Failed to create client", err)
		os.Exit(1)
	}

	if !c.IsLoggedIn() {
		printError("Session expired", fmt.Errorf("please login again with: tcrs login <user> <pass>"))
		os.Exit(1)
	}

	// Calculate week start date (Monday) if not specified
	date := saveDate
	if date == "" {
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		date = monday.Format("2006-01-02")
	}

	// Read input
	var input SaveInput
	var reader io.Reader

	if saveFile == "-" {
		reader = bufio.NewReader(os.Stdin)
		if IsVerbose() {
			fmt.Println("Reading from stdin...")
		}
	} else {
		file, err := os.Open(saveFile)
		if err != nil {
			printError("Failed to open file", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
		if IsVerbose() {
			fmt.Printf("Reading from %s...\n", saveFile)
		}
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		printError("Failed to read input", err)
		os.Exit(1)
	}

	if err := json.Unmarshal(data, &input); err != nil {
		printError("Failed to parse JSON", err)
		os.Exit(1)
	}

	if len(input.Entries) == 0 {
		printError("No entries to save", fmt.Errorf("entries array is empty"))
		os.Exit(1)
	}

	if IsVerbose() {
		fmt.Printf("Saving %d entries for week starting %s...\n", len(input.Entries), date)
	}

	err = c.SaveWeekTimecard(date, input.Entries)
	if err != nil {
		printError("Failed to save timecard", err)
		os.Exit(1)
	}

	if IsJSON() {
		result := map[string]interface{}{
			"success":         true,
			"week_start_date": date,
			"entries_saved":   len(input.Entries),
			"message":         "Timecard saved successfully",
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Successfully saved %d entries for week starting %s\n", len(input.Entries), date)
	}
}
