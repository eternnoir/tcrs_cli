package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
)

var weekDate string

var weekCmd = &cobra.Command{
	Use:   "week",
	Short: "View week timecard",
	Long:  `View the timecard entries for a specific week.`,
	Run:   runWeek,
}

func init() {
	rootCmd.AddCommand(weekCmd)
	weekCmd.Flags().StringVar(&weekDate, "date", "", "week start date in YYYY-MM-DD format (default: this week's Monday)")
}

func runWeek(cmd *cobra.Command, args []string) {
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
	date := weekDate
	if date == "" {
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		date = monday.Format("2006-01-02")
	}

	if IsVerbose() {
		fmt.Printf("Fetching week timecard for %s...\n", date)
	}

	result, err := c.GetWeekTimecard(date)
	if err != nil {
		printError("Failed to get week timecard", err)
		os.Exit(1)
	}

	if IsJSON() {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		printWeekTimecard(result)
	}
}

func printWeekTimecard(tc *client.WeekTimecard) {
	// Parse week start date
	startDate, err := time.Parse("2006-01-02", tc.WeekStartDate)
	if err != nil {
		startDate = time.Now()
	}

	// Print header
	fmt.Printf("Week Timecard: %s\n", tc.WeekStartDate)
	fmt.Println()

	// Day headers
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dates := make([]string, 7)
	for i := 0; i < 7; i++ {
		d := startDate.AddDate(0, 0, i)
		dates[i] = d.Format("01/02")
	}

	// Print table header
	fmt.Printf("%-30s", "Project/Activity")
	for i, day := range days {
		fmt.Printf(" %s(%s)", day, dates[i])
	}
	fmt.Println()

	// Print separator
	fmt.Print("------------------------------")
	for range days {
		fmt.Print("-----------")
	}
	fmt.Println()

	if len(tc.Entries) == 0 {
		fmt.Println("No entries")
	} else {
		for _, entry := range tc.Entries {
			// Truncate project name if too long
			name := entry.ProjectName
			if len(name) > 28 {
				name = name[:25] + "..."
			}
			fmt.Printf("%-30s", name)

			for _, day := range entry.Days {
				hoursStr := "   -   "
				switch v := day.Hours.(type) {
				case float64:
					if v > 0 {
						hoursStr = fmt.Sprintf("%7.1f", v)
					}
				case int:
					if v > 0 {
						hoursStr = fmt.Sprintf("%7d", v)
					}
				case string:
					if v != "" {
						hoursStr = fmt.Sprintf("%7s", v)
					}
				}
				fmt.Printf(" %s   ", hoursStr)
			}
			fmt.Println()
		}
	}

	// Print totals
	fmt.Print("------------------------------")
	for range days {
		fmt.Print("-----------")
	}
	fmt.Println()

	fmt.Printf("%-30s", "Total")
	var weekTotal float64
	for _, total := range tc.DailyTotals {
		weekTotal += total
		if total > 0 {
			fmt.Printf(" %7.1f   ", total)
		} else {
			fmt.Printf("    -      ")
		}
	}
	fmt.Println()

	fmt.Printf("\nWeek Total: %.1f hours\n", weekTotal)
}
