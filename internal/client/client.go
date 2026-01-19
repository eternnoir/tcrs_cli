package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/user/tcrs/internal/config"
)

// Client is the TCRS HTTP client.
type Client struct {
	cfg            *config.Config
	httpClient     *http.Client
	sessionManager *SessionManager
	userID         string
	loggedIn       bool
}

// NewClient creates a new TCRS client.
func NewClient(userID string, cfg *config.Config) (*Client, error) {
	// Validate base URL is configured
	if err := cfg.ValidateBaseURL(); err != nil {
		return nil, err
	}

	sm, err := NewSessionManager(userID, cfg)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Jar:     sm.CookieJar(),
		Timeout: 30 * time.Second,
	}

	c := &Client{
		cfg:            cfg,
		httpClient:     httpClient,
		sessionManager: sm,
		userID:         userID,
		loggedIn:       sm.HasValidSession(),
	}

	return c, nil
}

// setCommonHeaders sets common HTTP headers.
func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-TW,zh;q=0.9,en;q=0.8")
}

// Login authenticates with TCRS.
func (c *Client) Login(password string) error {
	if c.loggedIn {
		return nil
	}

	// First, get the login page
	loginPageURL := c.cfg.BaseURL + "/login.jsp"
	req, err := http.NewRequest("GET", loginPageURL, nil)
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)

	_, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get login page: %w", err)
	}

	// Perform login
	loginURL := c.cfg.BaseURL + "/servlet/VerifController"
	data := url.Values{
		"method": {"login"},
		"name":   {c.userID},
		"pw":     {password},
	}

	req, err = http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginPageURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyStr := string(body)

	// Check for login failure indicators
	bodyLower := strings.ToLower(bodyStr)
	if strings.Contains(bodyLower, "login failed") || strings.Contains(bodyLower, "invalid") {
		return ErrInvalidCredentials
	}

	// Verify by accessing a protected page
	verifyURL := c.cfg.BaseURL + "/Timecard/timecard_week/daychoose.jsp"
	req, err = http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 && !strings.Contains(strings.ToLower(resp.Request.URL.String()), "login") {
		c.loggedIn = true
		if err := c.sessionManager.SaveCookies(); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: could not save session cookies: %v\n", err)
		}
		return nil
	}

	return ErrLoginFailed
}

// Logout logs out from TCRS.
func (c *Client) Logout() error {
	if !c.loggedIn {
		return nil
	}

	logoutURL := c.cfg.BaseURL + "/servlet/VerifController?method=logout"
	req, err := http.NewRequest("GET", logoutURL, nil)
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)

	_, _ = c.httpClient.Do(req) // Ignore errors

	c.loggedIn = false
	return c.sessionManager.ClearCookies()
}

// IsLoggedIn returns true if the client is logged in.
func (c *Client) IsLoggedIn() bool {
	return c.loggedIn
}

// GetSessionInfo returns the current session info.
func (c *Client) GetSessionInfo() (*SessionInfo, error) {
	return c.sessionManager.GetSessionInfo()
}

// GetProjectsAndActivities retrieves projects and activities for a date.
func (c *Client) GetProjectsAndActivities(date string) (*ProjectsAndActivities, error) {
	if !c.loggedIn {
		return nil, ErrNotLoggedIn
	}

	activitiesURL := c.cfg.BaseURL + "/Timecard/timecard_week/daychoose.jsp?cho_date=" + url.QueryEscape(date)
	req, err := http.NewRequest("GET", activitiesURL, nil)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseProjectsAndActivities(string(body), date), nil
}

// GetWeekTimecard retrieves the week timecard for a given start date.
func (c *Client) GetWeekTimecard(weekStartDate string) (*WeekTimecard, error) {
	if !c.loggedIn {
		return nil, ErrNotLoggedIn
	}

	weekURL := c.cfg.BaseURL + "/Timecard/timecard_week/daychoose.jsp?cho_date=" + url.QueryEscape(weekStartDate)
	req, err := http.NewRequest("GET", weekURL, nil)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get week timecard: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseWeekTimecard(string(body), weekStartDate), nil
}

// SaveEntry represents an entry to save.
type SaveEntry struct {
	ProjectID  string         `json:"project_id"`
	ActivityID string         `json:"activity_id"`
	Progress   int            `json:"progress"`
	Days       []SaveDayEntry `json:"days"`
}

// SaveDayEntry represents a day entry to save.
type SaveDayEntry struct {
	Hours    interface{} `json:"hours"` // Can be float64 or empty string
	Note     string      `json:"note"`
	Progress int         `json:"progress"`
}

// SaveWeekTimecard saves timecard entries for a week.
func (c *Client) SaveWeekTimecard(weekStartDate string, entries []SaveEntry) error {
	if !c.loggedIn {
		return ErrNotLoggedIn
	}

	// First get projects to ensure we have the latest data
	_, err := c.GetProjectsAndActivities(weekStartDate)
	if err != nil {
		return fmt.Errorf("failed to get projects before save: %w", err)
	}

	saveURL := c.cfg.BaseURL + "/Timecard/timecard_week/weekinfo_deal.jsp"

	// Build form parameters
	params := make(map[string]string)
	dailyTotals := make([]float64, 7)

	// Process each entry
	for idx, entry := range entries {
		if entry.ProjectID == "" {
			continue
		}

		params[fmt.Sprintf("project%d", idx)] = entry.ProjectID

		// Build activity data: true$activity_id$project_id$0
		activityID := entry.ActivityID
		if activityID == "" {
			activityID = "xx"
		}
		activityData := fmt.Sprintf("true$%s$%s$0", activityID, entry.ProjectID)
		params[fmt.Sprintf("activity%d", idx)] = activityData
		params[fmt.Sprintf("actprogress%d", idx)] = strconv.Itoa(entry.Progress)

		// Process each day
		for dayIdx := 0; dayIdx < 7 && dayIdx < len(entry.Days); dayIdx++ {
			day := entry.Days[dayIdx]

			// Hours
			hoursStr := ""
			switch v := day.Hours.(type) {
			case float64:
				hoursStr = strconv.FormatFloat(v, 'f', -1, 64)
				dailyTotals[dayIdx] += v
			case int:
				hoursStr = strconv.Itoa(v)
				dailyTotals[dayIdx] += float64(v)
			case string:
				hoursStr = v
				if h, err := strconv.ParseFloat(v, 64); err == nil {
					dailyTotals[dayIdx] += h
				}
			}
			params[fmt.Sprintf("record%d_%d", idx, dayIdx)] = hoursStr
			params[fmt.Sprintf("note%d_%d", idx, dayIdx)] = day.Note
			params[fmt.Sprintf("progress%d_%d", idx, dayIdx)] = strconv.Itoa(day.Progress)
		}
	}

	// Fill empty entries (up to 25 projects)
	for emptyIdx := len(entries); emptyIdx < 25; emptyIdx++ {
		params[fmt.Sprintf("project%d", emptyIdx)] = ""
		params[fmt.Sprintf("activity%d", emptyIdx)] = ""
		params[fmt.Sprintf("actprogress%d", emptyIdx)] = ""
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			params[fmt.Sprintf("note%d_%d", emptyIdx, dayIdx)] = ""
			params[fmt.Sprintf("progress%d_%d", emptyIdx, dayIdx)] = ""
		}
	}

	// Add daily totals
	for dayIdx, total := range dailyTotals {
		params[fmt.Sprintf("norTotal%d", dayIdx)] = strconv.FormatFloat(total, 'f', -1, 64)
	}

	// Add overtime entries (all zeros)
	for idx := 0; idx < 25; idx++ {
		params[fmt.Sprintf("overactprogress%d", idx)] = "0"
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			params[fmt.Sprintf("overrecord%d_%d", idx, dayIdx)] = ""
			params[fmt.Sprintf("overnote%d_%d", idx, dayIdx)] = ""
			params[fmt.Sprintf("overprogress%d_%d", idx, dayIdx)] = "0"
		}
	}

	// Add overtime totals
	for dayIdx := 0; dayIdx < 7; dayIdx++ {
		params[fmt.Sprintf("oveTotal%d", dayIdx)] = "0"
	}

	// Build form data with specific order (mimicking browser behavior)
	formParts := []string{
		"save2=" + url.QueryEscape(" save "),
		"caller=this_week",
		"cdate=" + url.QueryEscape(weekStartDate),
	}

	// Add project/activity/record/note/progress params (sorted)
	projectKeys := make([]string, 0)
	for key := range params {
		if strings.HasPrefix(key, "project") || strings.HasPrefix(key, "activity") ||
			strings.HasPrefix(key, "actprogress") || strings.HasPrefix(key, "record") ||
			strings.HasPrefix(key, "note") || strings.HasPrefix(key, "progress") {
			if !strings.HasPrefix(key, "norTotal") && !strings.HasPrefix(key, "over") {
				projectKeys = append(projectKeys, key)
			}
		}
	}
	sort.Strings(projectKeys)
	for _, key := range projectKeys {
		formParts = append(formParts, key+"="+url.QueryEscape(params[key]))
	}

	// Add norTotal params
	for dayIdx := 0; dayIdx < 7; dayIdx++ {
		key := fmt.Sprintf("norTotal%d", dayIdx)
		formParts = append(formParts, key+"="+url.QueryEscape(params[key]))
	}

	// Add second caller param
	formParts = append(formParts, "caller=this_week")

	// Add overtime params (sorted)
	overKeys := make([]string, 0)
	for key := range params {
		if strings.HasPrefix(key, "over") || strings.HasPrefix(key, "ove") {
			overKeys = append(overKeys, key)
		}
	}
	sort.Strings(overKeys)
	for _, key := range overKeys {
		formParts = append(formParts, key+"="+url.QueryEscape(params[key]))
	}

	formData := strings.Join(formParts, "&")

	req, err := http.NewRequest("POST", saveURL, strings.NewReader(formData))
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.cfg.BaseURL+"/Timecard/timecard_week/daychoose.jsp?cho_date="+url.QueryEscape(weekStartDate))
	req.Header.Set("Origin", c.cfg.BaseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("save request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyLower := strings.ToLower(string(body))
	if strings.Contains(bodyLower, "error") || strings.Contains(bodyLower, "failed") {
		return fmt.Errorf("server indicated save failure")
	}

	return nil
}

// GetUserID returns the user ID.
func (c *Client) GetUserID() string {
	return c.userID
}
