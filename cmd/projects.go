package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tcrs/internal/client"
)

var projectsDate string

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects and activities",
	Long:  `List available projects and their activities for a given date.`,
	Run:   runProjects,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.Flags().StringVar(&projectsDate, "date", "", "date in YYYY-MM-DD format (default: today)")
}

func runProjects(cmd *cobra.Command, args []string) {
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

	// Use today's date if not specified
	date := projectsDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	if IsVerbose() {
		fmt.Printf("Fetching projects for %s...\n", date)
	}

	result, err := c.GetProjectsAndActivities(date)
	if err != nil {
		printError("Failed to get projects", err)
		os.Exit(1)
	}

	if IsJSON() {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Projects for %s:\n", result.Date)
		fmt.Println()

		if len(result.Projects) == 0 {
			fmt.Println("No projects found")
			return
		}

		for _, proj := range result.Projects {
			fmt.Printf("Project: %s (ID: %s)\n", proj.Name, proj.ID)

			if len(proj.Activities) == 0 {
				fmt.Println("  No activities")
			} else {
				for _, act := range proj.Activities {
					indent := ""
					if act.IndentLevel > 0 {
						indent = "    "
					}
					bottomMark := ""
					if act.IsBottom {
						bottomMark = " [leaf]"
					}
					fmt.Printf("  %s- %s (ID: %s)%s\n", indent, act.Name, act.ID, bottomMark)
				}
			}
			fmt.Println()
		}
	}
}
