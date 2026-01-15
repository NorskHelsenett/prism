package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"prism/config"
	"prism/database"
	"testing"
	"time"
)

var (
	baseURL     = getEnvOrDefault("PRISM_API_URL", "http://localhost:8080")
	oidcURL     = getEnvOrDefault("OIDC_URL", "http://host.containers.internal:9999")
	fixturesDir = getEnvOrDefault("FIXTURES_DIR", "./fixtures")
	// seeded API keys for tests: email -> apikey
	seededAPIKeys map[string]string
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestResult represents the outcome of a single access test
type TestResult struct {
	VulnerabilityID uint   `json:"vulnerability_id"`
	UserEmail       string `json:"user_email"`
	Expected        bool   `json:"expected"`
	Actual          bool   `json:"actual"`
	Passed          bool   `json:"passed"`
	Reason          string `json:"reason"`
	TestCategory    string `json:"test_category"`
	ResponseStatus  int    `json:"response_status"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

// TestReport contains all test results and summary statistics
type TestReport struct {
	Timestamp     time.Time                `json:"timestamp"`
	TotalTests    int                      `json:"total_tests"`
	Passed        int                      `json:"passed"`
	Failed        int                      `json:"failed"`
	PassRate      float64                  `json:"pass_rate"`
	Results       []TestResult             `json:"results"`
	FailedTests   []TestResult             `json:"failed_tests"`
	CategoryStats map[string]CategoryStats `json:"category_stats"`
}

// CategoryStats tracks pass/fail for each test category
type CategoryStats struct {
	Total  int     `json:"total"`
	Passed int     `json:"passed"`
	Failed int     `json:"failed"`
	Rate   float64 `json:"pass_rate"`
}

// authenticateUser logs in a user and returns the session cookie
// authenticateUser logs in a user and returns the access token (legacy behavior)
func authenticateUser(email, password string) (string, error) {
	// Preserve existing login endpoint behavior for tests that expect a token
	loginURL := fmt.Sprintf("%s/login", oidcURL)

	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login payload: %w", err)
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	return loginResp.AccessToken, nil
}

// authenticateUserCookieClient performs the full OIDC cookie-based flow and
// returns an http.Client which contains the session cookie for subsequent API calls.
func authenticateUserCookieClient(email, password string) (*http.Client, error) {
	// Create a client with a cookie jar to persist cookies across redirects
	jar, _ := cookiejar.New(nil)
	// Prevent automatic redirect following so we can capture the provider's redirect Location
	client := &http.Client{
		Jar:     jar,
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1) Start login on the Prism API to set state and code_verifier cookies
	loginInitiate := fmt.Sprintf("%s/api/login?provider=mock", baseURL)
	req, err := http.NewRequest("GET", loginInitiate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	// Do not follow redirects automatically here; let client follow but we need to capture redirect URL
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("initial login request failed: %w", err)
	}
	resp.Body.Close()

	// The last redirect should go to the OIDC provider authorize endpoint.
	// Build authorize URL by calling the OIDC provider /authorize with login param to pick user.
	// The OIDC mock accepts a login query parameter so we can simulate the user selecting account.

	// The initial /api/login response redirected to the provider's authorize URL and set the
	// server-side state/code_verifier cookies. Capture that Location (auth URL) so we use the
	// exact same state value the server stored in the cookie.
	prismLocation := resp.Header.Get("Location")
	if prismLocation == "" {
		return nil, fmt.Errorf("initial login response did not include Location header")
	}

	// Build the provider authorize URL by adding the login selector to the Location returned by Prism.
	authURL, err := url.Parse(prismLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Prism Location: %w", err)
	}
	q := authURL.Query()
	q.Set("login", email)
	authURL.RawQuery = q.Encode()

	// Call authorize URL (we expect a 302 Location back to redirect_uri containing code & state)
	authReq, _ := http.NewRequest("GET", authURL.String(), nil)
	authResp, err := client.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("authorize request failed: %w", err)
	}
	defer authResp.Body.Close()

	// Expect a redirect with Location header to the configured redirect_uri containing code & state
	loc := authResp.Header.Get("Location")
	if loc == "" {
		return nil, fmt.Errorf("authorize response did not include Location header")
	}

	// Parse Location to extract code and state
	parsed, err := url.Parse(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Location header: %w", err)
	}
	code := parsed.Query().Get("code")
	state := parsed.Query().Get("state")
	if code == "" {
		return nil, fmt.Errorf("authorize redirect missing code")
	}

	// Instead of following the redirect to whatever host is in the redirect_uri (which may be unreachable
	// from the test runner), call the Prism API callback endpoint at baseURL so the server can complete the
	// exchange using the cookies we set earlier.
	callbackURL := fmt.Sprintf("%s/api/callback?code=%s&state=%s", baseURL, url.QueryEscape(code), url.QueryEscape(state))
	cbReq, _ := http.NewRequest("GET", callbackURL, nil)
	cbResp, err := client.Do(cbReq)
	if err != nil {
		return nil, fmt.Errorf("callback request to Prism failed: %w", err)
	}
	defer cbResp.Body.Close()

	// Now the Prism server should have created a session and set the session cookie in the jar.
	// Verify by attempting to hit an authenticated endpoint on Prism API using this client
	testURL := fmt.Sprintf("%s/api/vulnerability/1", baseURL)
	testReq, _ := http.NewRequest("GET", testURL, nil)
	testResp, err := client.Do(testReq)
	if err != nil {
		return nil, fmt.Errorf("session validation request failed: %w", err)
	}
	defer testResp.Body.Close()

	if testResp.StatusCode == http.StatusUnauthorized || testResp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(testResp.Body)
		return nil, fmt.Errorf("cookie-based authentication did not produce a valid session: status=%d body=%s", testResp.StatusCode, string(body))
	}

	return client, nil
}

// testVulnerabilityAccess attempts to access a vulnerability as a specific user
func testVulnerabilityAccess(vulnID uint, userEmail, password string) (bool, int, error) {
	// If we have a seeded API key for this user, use it (fast and reliable for integration tests)
	url := fmt.Sprintf("%s/api/vulnerability/%d", baseURL, vulnID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if seededAPIKeys != nil {
		if key, ok := seededAPIKeys[userEmail]; ok && key != "" {
			req.Header.Set("x-api-key", key)
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				_, _ = io.ReadAll(resp.Body)
				hasAccess := resp.StatusCode == http.StatusOK
				return hasAccess, resp.StatusCode, nil
			}
			// if API key request failed, fall through to cookie/token flows
		}
	}

	// First attempt: cookie-based flow
	client, cerr := authenticateUserCookieClient(userEmail, password)
	if cerr == nil && client != nil {
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
			hasAccess := resp.StatusCode == http.StatusOK
			return hasAccess, resp.StatusCode, nil
		}
		// if cookie client failed to perform request, fall through to token flow
	}

	// Fallback: token-based flow (legacy behavior)
	token, terr := authenticateUser(userEmail, password)
	if terr != nil {
		// return the more informative error(s)
		if cerr != nil {
			return false, 0, fmt.Errorf("cookie error: %v; token error: %v", cerr, terr)
		}
		return false, 0, fmt.Errorf("token authentication failed: %w", terr)
	}

	// Attempt to access vulnerability using the bearer token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client2 := &http.Client{Timeout: 10 * time.Second}
	resp, err := client2.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	hasAccess := resp.StatusCode == http.StatusOK
	return hasAccess, resp.StatusCode, nil
}

// TestAccessMatrix runs the comprehensive access control test suite
func TestAccessMatrix(t *testing.T) {
	// Set default config path if not already set
	if os.Getenv("CONFIG_PATH") == "" {
		cwd, _ := os.Getwd()
		os.Setenv("CONFIG_PATH", filepath.Join(cwd, "config.test.yaml"))
	}
	
	// Set default database path if not already set
	if os.Getenv("DATABASE_PATH") == "" {
		cwd, _ := os.Getwd()
		os.Setenv("DATABASE_PATH", filepath.Join(cwd, "test-data")+"/")
	}
	
	// Set default roles path if not already set
	if os.Getenv("ROLES_PATH") == "" {
		cwd, _ := os.Getwd()
		os.Setenv("ROLES_PATH", filepath.Join(cwd, "../../api/roles.yaml"))
	}
	
	// Wait for services to be ready
	if err := waitForServices(); err != nil {
		t.Fatalf("Services not ready: %v", err)
	}

	// Load access matrix
	matrixFile := filepath.Join(fixturesDir, "access_matrix.csv")
	matrix, err := LoadAccessMatrix(matrixFile)
	if err != nil {
		t.Fatalf("Failed to load access matrix: %v", err)
	}

	// Ensure config is loaded before initializing DB
	if err := func() error {
		if config.AppConfig == nil {
			if err := config.LoadConfig(); err != nil {
				return fmt.Errorf("Failed to load config: %w", err)
			}
		}

		// Clean existing test database to ensure fresh state
		dbPath := filepath.Join(os.Getenv("DATABASE_PATH"), "prism.db")
		os.Remove(dbPath)

		// Initialize database
		database.InitDB()

		// Seed users first
		usersFile := filepath.Join(fixturesDir, "users.csv")
		if err := SeedUsers(usersFile); err != nil {
			return fmt.Errorf("Failed to seed users: %w", err)
		}

		// Always seed the database with fixture data before running tests
		projectsFile := filepath.Join(fixturesDir, "projects.csv")
		vulnsFile := filepath.Join(fixturesDir, "vulnerabilities.csv")
		if err := SeedDatabase(projectsFile, vulnsFile); err != nil {
			t.Fatalf("Failed to seed database: %v", err)
		}

		return nil
	}(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	var users []string
	for _, e := range matrix {
		users = append(users, e.UserEmail)
	}
	seededAPIKeys, err = SeedAPIKeysForUsers(users)
	if err != nil {
		t.Fatalf("Failed to seed API keys: %v", err)
	}

	t.Logf("Running %d access control tests...", len(matrix))

	var results []TestResult
	categoryStats := make(map[string]*CategoryStats)

	for i, entry := range matrix {
		// Get user password (all test users use password123)
		password := "password123"

		// Test access
		hasAccess, statusCode, err := testVulnerabilityAccess(
			entry.VulnerabilityID,
			entry.UserEmail,
			password,
		)

		result := TestResult{
			VulnerabilityID: entry.VulnerabilityID,
			UserEmail:       entry.UserEmail,
			Expected:        entry.ShouldAccess,
			Actual:          hasAccess,
			Passed:          hasAccess == entry.ShouldAccess,
			Reason:          entry.Reason,
			TestCategory:    entry.TestCategory,
			ResponseStatus:  statusCode,
		}

		if err != nil {
			result.ErrorMessage = err.Error()
			result.Passed = false
		}

		results = append(results, result)

		// Update category stats
		if categoryStats[entry.TestCategory] == nil {
			categoryStats[entry.TestCategory] = &CategoryStats{}
		}
		categoryStats[entry.TestCategory].Total++
		if result.Passed {
			categoryStats[entry.TestCategory].Passed++
		} else {
			categoryStats[entry.TestCategory].Failed++
		}

		// Log progress
		if (i+1)%10 == 0 || !result.Passed {
			status := "✓"
			if !result.Passed {
				status = "✗"
			}
			t.Logf("[%d/%d] %s vuln=%d, user=%s, expect=%v, got=%v (%s)",
				i+1, len(matrix), status,
				entry.VulnerabilityID, entry.UserEmail,
				entry.ShouldAccess, hasAccess, entry.Reason)
		}
	}

	// Generate report
	report := generateReport(results, categoryStats)

	// Save report to file
	if err := saveReport(report); err != nil {
		t.Errorf("Failed to save report: %v", err)
	}

	// Print summary
	t.Logf("\n=== Test Summary ===")
	t.Logf("Total Tests: %d", report.TotalTests)
	t.Logf("Passed: %d", report.Passed)
	t.Logf("Failed: %d", report.Failed)
	t.Logf("Pass Rate: %.2f%%", report.PassRate)

	t.Logf("\n=== Category Breakdown ===")
	for category, stats := range report.CategoryStats {
		t.Logf("%s: %d/%d (%.1f%%)", category, stats.Passed, stats.Total, stats.Rate)
	}

	if report.Failed > 0 {
		t.Logf("\n=== Failed Tests ===")
		for _, failed := range report.FailedTests {
			t.Logf("✗ vuln=%d, user=%s, expected=%v, got=%v (%s)",
				failed.VulnerabilityID, failed.UserEmail,
				failed.Expected, failed.Actual, failed.Reason)
			if failed.ErrorMessage != "" {
				t.Logf("  Error: %s", failed.ErrorMessage)
			}
		}
		t.Errorf("%d tests failed", report.Failed)
	}
}

func generateReport(results []TestResult, categoryStats map[string]*CategoryStats) *TestReport {
	report := &TestReport{
		Timestamp:     time.Now(),
		Results:       results,
		CategoryStats: make(map[string]CategoryStats),
	}

	for _, r := range results {
		report.TotalTests++
		if r.Passed {
			report.Passed++
		} else {
			report.Failed++
			report.FailedTests = append(report.FailedTests, r)
		}
	}

	if report.TotalTests > 0 {
		report.PassRate = float64(report.Passed) / float64(report.TotalTests) * 100
	}

	// Calculate category stats
	for category, stats := range categoryStats {
		if stats.Total > 0 {
			stats.Rate = float64(stats.Passed) / float64(stats.Total) * 100
		}
		report.CategoryStats[category] = *stats
	}

	return report
}

func saveReport(report *TestReport) error {
	reportsDir := "./reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	filename := filepath.Join(reportsDir, fmt.Sprintf("access_test_%s.json",
		report.Timestamp.Format("20060102_150405")))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Also save as latest
	latestFile := filepath.Join(reportsDir, "access_test_latest.json")
	if err := os.WriteFile(latestFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write latest report: %w", err)
	}

	return nil
}

func waitForServices() error {
	maxRetries := 30
	retryDelay := 2 * time.Second

	// Wait for API (use health endpoint on port 8888)
	for i := 0; i < maxRetries; i++ {
		healthURL := fmt.Sprintf("%s:8888/healthz", baseURL[:len(baseURL)-5]) // Remove :8080 and add :8888/healthz
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("API not ready after %d retries", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	// Wait for OIDC mock
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/", oidcURL))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if i == maxRetries-1 {
			return fmt.Errorf("OIDC mock not ready after %d retries", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	return nil
}
