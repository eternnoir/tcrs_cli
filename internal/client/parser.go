package client

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Project represents a TCRS project.
type Project struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Activities []Activity `json:"activities"`
}

// Activity represents an activity within a project.
type Activity struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	IsBottom    bool   `json:"is_bottom"`
	UID         string `json:"uid"`
	Progress    string `json:"progress"`
	IndentLevel int    `json:"indent_level"`
}

// DayEntry represents a single day's timecard entry.
type DayEntry struct {
	Hours    interface{} `json:"hours"` // Can be float64 or empty string
	Note     string      `json:"note"`
	Progress int         `json:"progress"`
}

// WeekEntry represents a week's timecard entry for a project.
type WeekEntry struct {
	ProjectID    string     `json:"project_id"`
	ProjectName  string     `json:"project_name"`
	ActivityData string     `json:"activity_data"`
	Progress     int        `json:"progress"`
	Days         []DayEntry `json:"days"`
}

// WeekTimecard represents a full week's timecard data.
type WeekTimecard struct {
	WeekStartDate string      `json:"week_start_date"`
	Entries       []WeekEntry `json:"entries"`
	DailyTotals   []float64   `json:"daily_totals"`
}

// ProjectsAndActivities represents the result of parsing projects and activities.
type ProjectsAndActivities struct {
	Date     string    `json:"date"`
	Projects []Project `json:"projects"`
}

// ParseProjectsAndActivities parses HTML content to extract projects and activities.
func ParseProjectsAndActivities(htmlContent, date string) *ProjectsAndActivities {
	projects, activities := parseJSArrays(htmlContent)

	result := &ProjectsAndActivities{
		Date:     date,
		Projects: make([]Project, 0),
	}

	// Create project map with activities
	projectMap := make(map[string]*Project)
	for _, p := range projects {
		proj := Project{
			ID:         p.ID,
			Name:       p.Name,
			Activities: make([]Activity, 0),
		}
		projectMap[p.ID] = &proj
	}

	// Add activities to their respective projects
	for _, act := range activities {
		if proj, ok := projectMap[act.ProjectID]; ok {
			// Use UID as ID for the activity
			actCopy := act
			actCopy.ID = act.UID
			proj.Activities = append(proj.Activities, actCopy)
		}
	}

	// Convert map to slice
	for _, proj := range projectMap {
		result.Projects = append(result.Projects, *proj)
	}

	return result
}

// parseJSArrays extracts projects and activities from JavaScript in HTML.
func parseJSArrays(htmlContent string) ([]Project, []Activity) {
	projects := make([]Project, 0)
	activities := make([]Activity, 0)
	projectMap := make(map[string]string) // id -> name

	// Pattern 1: Standard dropdown: <option value="PROJECT_ID">PROJECT_NAME</option>
	projectPattern1 := regexp.MustCompile(`<option value="(\d+)">([^<]+)</option>`)
	matches1 := projectPattern1.FindAllStringSubmatch(htmlContent, -1)

	// Pattern 2: Dropdown with attributes: <option value="PROJECT_ID" ...>PROJECT_NAME</option>
	projectPattern2 := regexp.MustCompile(`<option[^>]*\svalue="(\d+)"[^>]*>([^<]+)</option>`)
	matches2 := projectPattern2.FindAllStringSubmatch(htmlContent, -1)

	// Combine all project matches
	seen := make(map[string]bool)
	for _, match := range append(matches1, matches2...) {
		if len(match) >= 3 {
			value := match[1]
			name := strings.TrimSpace(match[2])

			// Skip default options and special projects
			if strings.Contains(strings.ToLower(name), "select project") || value == "--" || name == "" {
				continue
			}

			key := value + "_" + name
			if !seen[key] {
				seen[key] = true
				projectMap[value] = name
				projects = append(projects, Project{
					ID:   value,
					Name: name,
				})
			}
		}
	}

	// Extract activities from JavaScript: act.append('PROJECT_ID','ACTIVITY_NAME','IS_BOTTOM','UID','PROGRESS')
	activityPattern := regexp.MustCompile(`act\.append\('(\d+)',\s*'([^']+)',\s*'([^']+)',\s*'([^']+)',\s*'([^']+)'\)`)
	activityMatches := activityPattern.FindAllStringSubmatch(htmlContent, -1)

	// If no projects found from dropdown, extract from activities
	if len(projects) == 0 {
		projectIDs := make(map[string]bool)
		for _, match := range activityMatches {
			if len(match) >= 2 {
				projectIDs[match[1]] = true
			}
		}

		for projectID := range projectIDs {
			for _, match := range activityMatches {
				if len(match) >= 2 && match[1] == projectID {
					activityName := strings.TrimSpace(match[2])
					projectName := activityName

					// Extract prefix as project name if has <<x>> format
					nameMatch := regexp.MustCompile(`([^<]+)\s*<<`).FindStringSubmatch(activityName)
					if len(nameMatch) >= 2 {
						projectName = strings.TrimSpace(nameMatch[1])
					}

					projectMap[projectID] = projectName
					projects = append(projects, Project{
						ID:   projectID,
						Name: projectName,
					})
					break
				}
			}
		}
	}

	// Process all activities
	for _, match := range activityMatches {
		if len(match) >= 6 {
			projectID := match[1]
			activityName := match[2]
			isBottom := strings.ToLower(match[3]) == "true"
			uid := match[4]
			progress := match[5]

			// Ensure we have this project
			if _, ok := projectMap[projectID]; !ok {
				projectMap[projectID] = "Project " + projectID
				projects = append(projects, Project{
					ID:   projectID,
					Name: projectMap[projectID],
				})
			}

			// Process activity name for hierarchy info
			indentLevel := 0
			cleanedName := activityName

			// Check for leading spaces
			if match := regexp.MustCompile(`^\s+`).FindString(activityName); match != "" {
				indentLevel = len(match)
				cleanedName = strings.TrimSpace(activityName)
			}

			// Remove number prefix like "1. Activity Name"
			if match := regexp.MustCompile(`^(\d+[\.\)]\s+)(.*)`).FindStringSubmatch(cleanedName); len(match) >= 3 {
				cleanedName = match[2]
			}

			// Remove <<x.y.z>> markers
			if match := regexp.MustCompile(`(.+)\s*<<[^>]+>>`).FindStringSubmatch(cleanedName); len(match) >= 2 {
				cleanedName = strings.TrimSpace(match[1])
			}

			activities = append(activities, Activity{
				ID:          projectID + "_" + cleanedName + "_" + uid,
				ProjectID:   projectID,
				Name:        cleanedName,
				FullName:    activityName,
				IsBottom:    isBottom,
				UID:         uid,
				Progress:    progress,
				IndentLevel: indentLevel,
			})
		}
	}

	return projects, activities
}

// ParseWeekTimecard parses HTML content to extract week timecard data.
func ParseWeekTimecard(htmlContent, weekStartDate string) *WeekTimecard {
	result := &WeekTimecard{
		WeekStartDate: weekStartDate,
		Entries:       make([]WeekEntry, 0),
		DailyTotals:   []float64{0, 0, 0, 0, 0, 0, 0},
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return result
	}

	// Find table rows
	doc.Find("table.timecard_table tr").Each(func(i int, row *goquery.Selection) {
		// Check if this row has project selects
		projectSelect := row.Find("select[name^='project']")
		if projectSelect.Length() == 0 {
			return
		}

		// Extract row index from select name
		selectName, _ := projectSelect.Attr("name")
		if !strings.HasPrefix(selectName, "project") {
			return
		}
		idxStr := strings.TrimPrefix(selectName, "project")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return
		}

		// Get selected project
		var projectID, projectName string
		projectSelect.Find("option[selected]").Each(func(_ int, opt *goquery.Selection) {
			projectID, _ = opt.Attr("value")
			projectName = strings.TrimSpace(opt.Text())
		})

		// Skip empty projects
		if projectID == "" || projectID == "--" {
			return
		}

		// Get activity data
		var activityData string
		activitySelect := row.Find("select[name='activity" + idxStr + "']")
		activitySelect.Find("option[selected]").Each(func(_ int, opt *goquery.Selection) {
			activityData, _ = opt.Attr("value")
		})

		// Get progress
		progress := 0
		if progressInput := row.Find("input[name='actprogress" + idxStr + "']"); progressInput.Length() > 0 {
			if val, _ := progressInput.Attr("value"); val != "" {
				progress, _ = strconv.Atoi(val)
			}
		}

		// Extract day data
		days := make([]DayEntry, 7)
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayEntry := DayEntry{
				Hours:    "",
				Note:     "",
				Progress: 0,
			}

			// Get hours
			hourInput := row.Find("input[name='record" + strconv.Itoa(idx) + "_" + strconv.Itoa(dayIdx) + "']")
			if hourInput.Length() > 0 {
				if val, _ := hourInput.Attr("value"); val != "" && strings.TrimSpace(val) != "" {
					if hours, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
						dayEntry.Hours = hours
						result.DailyTotals[dayIdx] += hours
					}
				}
			}

			// Get note
			noteInput := row.Find("input[name='note" + strconv.Itoa(idx) + "_" + strconv.Itoa(dayIdx) + "']")
			if noteInput.Length() > 0 {
				dayEntry.Note, _ = noteInput.Attr("value")
			}

			// Get progress
			progressInput := row.Find("input[name='progress" + strconv.Itoa(idx) + "_" + strconv.Itoa(dayIdx) + "']")
			if progressInput.Length() > 0 {
				if val, _ := progressInput.Attr("value"); val != "" {
					dayEntry.Progress, _ = strconv.Atoi(val)
				}
			}

			days[dayIdx] = dayEntry
		}

		result.Entries = append(result.Entries, WeekEntry{
			ProjectID:    projectID,
			ProjectName:  projectName,
			ActivityData: activityData,
			Progress:     progress,
			Days:         days,
		})
	})

	// Try to get actual totals from subtotal row
	doc.Find("tr.subtotal td").Each(func(i int, td *goquery.Selection) {
		if i == 0 {
			return // Skip label cell
		}
		dayIdx := i - 1
		if dayIdx < 7 {
			val := strings.TrimSpace(td.Text())
			if total, err := strconv.ParseFloat(val, 64); err == nil {
				result.DailyTotals[dayIdx] = total
			}
		}
	})

	return result
}
