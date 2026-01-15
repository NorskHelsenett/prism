package integration

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"prism/config"
	"prism/crypto"
	"prism/database"
	"prism/models"

	"gorm.io/datatypes"
)

// TestUser represents a user from the CSV
type TestUser struct {
	Email       string
	Role        string
	Password    string
	SkipOTP     bool
	DisplayName string
}

// TestProject represents a project from the CSV
type TestProject struct {
	ID           uint
	Name         string
	Description  string
	SlackChannel string
	ClientEmail  string
	HackerName   string
	IsBugBounty  bool
}

// TestVulnerability represents a vulnerability from the CSV
type TestVulnerability struct {
	ID          uint
	Title       string
	Description string
	Severity    string
	CVSSScore   float64
	ProjectID   uint
	Visibility  string
	FoundBy     string
	AssignedTo  string
	Status      string
	Impact      string
	Likelihood  string
	Remediation string
}

// AccessMatrixEntry represents a single test case from access_matrix.csv
type AccessMatrixEntry struct {
	VulnerabilityID uint
	UserEmail       string
	ShouldAccess    bool
	Reason          string
	TestCategory    string
}

// LoadUsers parses users.csv
func LoadUsers(filePath string) ([]TestUser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open users file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var users []TestUser
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		skipOTP, _ := strconv.ParseBool(record[3])
		users = append(users, TestUser{
			Email:       strings.TrimSpace(record[0]),
			Role:        strings.TrimSpace(record[1]),
			Password:    strings.TrimSpace(record[2]),
			SkipOTP:     skipOTP,
			DisplayName: strings.TrimSpace(record[4]),
		})
	}

	return users, nil
}

func LoadAPIKeys() (map[string]string, error) {
	result := make(map[string]string)

	// Load all users from db
	users, err := database.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to load users from database: %w", err)
	}

	// Loop through users and load their API keys
	for _, user := range *users {
		apiKeys, err := database.GetApiKeys(user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get API keys for user %s: %w", user.Email, err)
		}
		if len(*apiKeys) > 0 {
			var keys []string
			for _, k := range *apiKeys {
				keys = append(keys, k.ApiKey)
			}
			result[user.Email] = strings.Join(keys, ",")
		} else {
			result[user.Email] = "" // No API key found
		}
	}

	return result, nil
}

// SeedUsers populates the test database with users from users.csv
func SeedUsers(usersFile string) error {
	users, err := LoadUsers(usersFile)
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}
	for _, u := range users {
		user := &database.UserData{
			Email: u.Email,
			Name:  u.DisplayName,
			Role:  u.Role,
			Title: "Test User",
		}
		if err := database.SaveOrUpdateUserData(user.Name, user.Email, ""); err != nil {
			return fmt.Errorf("failed to insert user %s: %w", u.Email, err)
		}
	}
	return nil
}

// LoadProjects parses projects.csv
func LoadProjects(filePath string) ([]TestProject, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open projects file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Helper to normalize fields that are intended to be comma-separated lists
	normalizeList := func(field string) string {
		// csv.Reader already handles quoted commas, but callers may expect
		// a clean comma-separated list without extra spaces. Split on comma
		// and re-join trimmed elements. If the field is empty, return empty.
		field = strings.TrimSpace(field)
		if field == "" {
			return ""
		}
		parts := strings.Split(field, ",")
		var out []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			out = append(out, p)
		}
		return strings.Join(out, ",")
	}

	var projects []TestProject
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		id, _ := strconv.ParseUint(record[0], 10, 32)
		isBugBounty, _ := strconv.ParseBool(record[6])

		projects = append(projects, TestProject{
			ID:           uint(id),
			Name:         strings.TrimSpace(record[1]),
			Description:  strings.TrimSpace(record[2]),
			SlackChannel: strings.TrimSpace(record[3]),
			ClientEmail:  normalizeList(record[4]),
			HackerName:   normalizeList(record[5]),
			IsBugBounty:  isBugBounty,
		})
	}

	return projects, nil
}

// LoadVulnerabilities parses vulnerabilities.csv
func LoadVulnerabilities(filePath string) ([]TestVulnerability, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open vulnerabilities file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var vulnerabilities []TestVulnerability
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		id, _ := strconv.ParseUint(record[0], 10, 32)
		cvss, _ := strconv.ParseFloat(record[4], 64)
		projectID, _ := strconv.ParseUint(record[5], 10, 32)

		vulnerabilities = append(vulnerabilities, TestVulnerability{
			ID:          uint(id),
			Title:       strings.TrimSpace(record[1]),
			Description: strings.TrimSpace(record[2]),
			Severity:    strings.TrimSpace(record[3]),
			CVSSScore:   cvss,
			ProjectID:   uint(projectID),
			Visibility:  strings.TrimSpace(record[6]),
			FoundBy:     strings.TrimSpace(record[7]),
			AssignedTo:  strings.TrimSpace(record[8]),
			Status:      strings.TrimSpace(record[9]),
			Impact:      strings.TrimSpace(record[10]),
			Likelihood:  strings.TrimSpace(record[11]),
			Remediation: strings.TrimSpace(record[12]),
		})
	}

	return vulnerabilities, nil
}

// LoadAccessMatrix parses access_matrix.csv
func LoadAccessMatrix(filePath string) ([]AccessMatrixEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open access matrix file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var entries []AccessMatrixEntry
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		vulnID, _ := strconv.ParseUint(record[0], 10, 32)
		shouldAccess, _ := strconv.ParseBool(record[2])

		entries = append(entries, AccessMatrixEntry{
			VulnerabilityID: uint(vulnID),
			UserEmail:       strings.TrimSpace(record[1]),
			ShouldAccess:    shouldAccess,
			Reason:          strings.TrimSpace(record[3]),
			TestCategory:    strings.TrimSpace(record[4]),
		})
	}

	return entries, nil
}

// SeedDatabase populates the test database with fixture data
func SeedDatabase(projectsFile, vulnsFile string) error {

	database.InitDB()

	// Load and seed projects
	projects, err := LoadProjects(projectsFile)
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	for _, p := range projects {
		project := &database.ProjectData{
			ProjectName:  p.Name,
			Description:  p.Description,
			SlackChannel: p.SlackChannel,
			ClientEmail:  p.ClientEmail,
			HackerName:   p.HackerName,
			IsBugBounty:  p.IsBugBounty,
		}
		project.ID = p.ID
		database.CreateProject(project)
	}

	// Load and seed vulnerabilities
	vulns, err := LoadVulnerabilities(vulnsFile)
	if err != nil {
		return fmt.Errorf("failed to load vulnerabilities: %w", err)
	}

	for _, v := range vulns {
		// Create vulnerability JSON payload
		vulnData := map[string]interface{}{
			"title":       v.Title,
			"description": v.Description,
			"severity":    v.Severity,
			"cvss_score":  v.CVSSScore,
			"visibility":  v.Visibility,
			"assignedTo":  v.AssignedTo,
			"status":      v.Status,
			"impact":      v.Impact,
			"likelihood":  v.Likelihood,
			"remediation": v.Remediation,
		}

		vulnJSON, err := json.Marshal(vulnData)
		if err != nil {
			return fmt.Errorf("failed to marshal vulnerability %d: %w", v.ID, err)
		}

		jsonData := &database.JSONData{
			Vulnerability: datatypes.JSON(vulnJSON),
			FoundBy:       v.FoundBy,
			ProjectID:     &v.ProjectID,
			Status:        v.Status,
		}
		jsonData.ID = v.ID

		if err := database.CreateJSONData(jsonData); err != nil {
			return fmt.Errorf("failed to create vulnerability %d: %w", v.ID, err)
		}
	}

	usersFile := filepath.Join(filepath.Dir(projectsFile), "users.csv")
	if err := SeedUsers(usersFile); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	users, err := LoadUsers(usersFile)
	if err != nil {
		return fmt.Errorf("failed to reload users for API key seeding: %w", err)
	}
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}
	_, err = SeedAPIKeysForUsers(emails)
	if err != nil {
		return fmt.Errorf("failed to seed API keys: %w", err)
	}

	return nil
}

// SeedAPIKeysForUsers creates an API key for each provided user email and persists it in the database.
// It returns a map from email -> plaintext API key that tests can use in the x-api-key header.
func SeedAPIKeysForUsers(emails []string) (map[string]string, error) {
	result := make(map[string]string)

	// Ensure configuration is loaded so crypto.HashAPIKey can access HMAC_SECRET_KEY
	if config.AppConfig == nil {
		if err := config.LoadConfig(); err != nil {
			return nil, fmt.Errorf("failed to load config for seeding apikeys: %w", err)
		}
	}

	// // Ensure the database package has initialized its DB connection so PersistApiKey
	// // can use the gorm DB pointer. InitDB will run migrations if necessary.
	// database.InitDB()

	for _, email := range emails {
		apimodel := &models.APIKey{
			Name:  "integration-test",
			Email: email,
		}

		created, err := crypto.CreateAPIKey(apimodel)
		if err != nil {
			return nil, fmt.Errorf("failed to create apikey for %s: %w", email, err)
		}

		persisted, err := database.PersistApiKey(created)
		if err != nil {
			return nil, fmt.Errorf("failed to persist apikey for %s: %w", email, err)
		}

		// created.ApiKey holds the plaintext key returned by CreateAPIKey
		result[persisted.Email] = created.ApiKey
	}

	return result, nil
}

// CleanDatabase removes all test data
// Note: This is a placeholder. In practice, the test database is recreated for each test run.
// If you need to clean specific data, implement using database.DeleteVulnerability() and
// database.DeleteProjectAndAssets() for individual records.
func CleanDatabase() error {
	// The database is typically ephemeral in tests and gets recreated
	// Individual cleanup can be done if needed using:
	// - database.DeleteVulnerability(id)
	// - database.DeleteProjectAndAssets(id)
	return nil
}
