package database

import (
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"os"
	"path/filepath"
	"time"

	"prism/config"
)

var (
	ErrNotFound = errors.New("record not found")
	// Other custom errors can be defined here
)

type ProjectData struct {
	gorm.Model
	ProjectName  string `gorm:"not null" binding:"required"`
	SlackChannel string
	Description  string
	ClientEmail  string
	HackerName   string
	IsBugBounty  bool
}

func CreateProject(project *ProjectData) {
	result := db.Create(project) // Create a new record
	if result.Error != nil {
		// Handle error here, for example:
		panic(result.Error)
	}
}

func UpdateProject(project *ProjectData) error {
	// Begin a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Step 1: Update regular fields, excluding the boolean field
	if err := tx.Model(&ProjectData{}).Where("id = ?", project.ID).
		Omit("IsBugBounty").Updates(project).Error; err != nil {
		tx.Rollback() // Rollback the transaction in case of error
		return err
	}

	// Step 2: Update the boolean field alone
	if err := tx.Model(&ProjectData{}).Where("id = ?", project.ID).
		Select("IsBugBounty").Updates(map[string]interface{}{
		"IsBugBounty": project.IsBugBounty,
	}).Error; err != nil {
		tx.Rollback() // Rollback the transaction in case of error
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

// JSONData is a simple model for storing JSON data
type JSONData struct {
	gorm.Model
	Vulnerability datatypes.JSON
	FoundBy       string
	ProjectID     *uint        // Foreign key for ProjectData
	Project       *ProjectData // The associated project
	Status        string       `gorm:"default:Reported"`
	SlackUrl      string
}

type SlackSettings struct {
	Enabled   bool   `json:"enabled"`
	ChannelID string `json:"channelID"`
	Workspace string `json:"workspace"`
}

type AuditLoggingSettings struct {
	Enabled bool `json:"enabled"`
}

type Settings struct {
	gorm.Model
	SlackData    string               `json:"-" gorm:"column:slack_settings"`
	Slack        SlackSettings        `gorm:"-" json:"slack"`
	AuditLogData string               `json:"-" gorm:"column:auditlog_settings"`
	AuditLog     AuditLoggingSettings `gorm:"-" json:"auditlog"`
}

type UserData struct {
	gorm.Model
	Email   string
	Name    string
	Picture string
	Role    string
	Title   string `gorm:"default:My title"`
}

type Vulnerability struct {
	Criticality string `json:"criticality"`
	Category    string `json:"category"`
	ProjectID   uint   `json:"projectID"`
}

type EventQueue struct {
	ID        uint `gorm:"primaryKey"`
	TableID   uint
	TableName string
	Processed bool      `gorm:"default:false;index:idx_processed"`
	CreatedAt time.Time `gorm:"index:idx_created_at,autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

type AuditLog struct {
	Timestamp   time.Time
	UserEmail   string
	Method      string
	Action      string
	Status      string
	Description string
}

var db *gorm.DB

func InitDB() {
	appConfig, _ := config.LoadConfig()
	var err error
	db, err = gorm.Open(sqlite.Open(appConfig.Database.Path+"/prism.db?cache=shared&_synchronous=FULL"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database object")
	}

	// Set the maximum number of open connections to the database
	// Since SQLite over NFS is not ideal for high concurrency, especially with writes,
	// we keep the max open connections low.
	sqlDB.SetMaxOpenConns(10)

	// Set the maximum number of idle connections to the database
	sqlDB.SetMaxIdleConns(2)

	// Set the maximum amount of time a connection may be reused
	// For a web API, it can be reasonable to have a short max lifetime to refresh connections regularly
	sqlDB.SetConnMaxLifetime(30 * time.Second) // 30 seconds

	// Enable WAL mode
	if err := db.Exec("PRAGMA journal_mode = WAL;").Error; err != nil {
		panic("failed to set journal_mode to WAL")
	}
	// Set cache size to 20000 pages. Each page is usually 4KB.
	if err := db.Exec("PRAGMA cache_size = 20000;").Error; err != nil {
		panic("failed to set cache_size")
	}
	if err := db.Exec("PRAGMA locking_mode = NORMAL;").Error; err != nil {
		panic("failed to set locking_mode")
	}

	// Migrate the schema
	db.AutoMigrate(&JSONData{})
	db.AutoMigrate(&UserData{})
	db.AutoMigrate(&ProjectData{})
	db.AutoMigrate(&EventQueue{})
	db.AutoMigrate(&Settings{})
	db.AutoMigrate(&AuditLog{})

	db.Exec(`
    CREATE TRIGGER IF NOT EXISTS jsondata_insert AFTER INSERT ON json_data
		BEGIN
    	INSERT INTO event_queues (table_id, table_name, created_at) VALUES (NEW.id, 'vulnerability', CURRENT_TIMESTAMP);
		END;
		`)
}

func RecordAuditLog(log AuditLog) error {
	return db.Create(&log).Error
}

func SetEventProcessed(event *EventQueue) {
	db.Model(&event).Update("processed", true)
}

func GetAllAudits(limit int) (*[]AuditLog, error) {
	var auditLog []AuditLog

	if limit <= 0 {
		limit = 50 // Default limit
	}

	result := db.Order("timestamp desc").Limit(limit).Find(&auditLog)

	if result.Error != nil {
		return nil, result.Error
	}

	return &auditLog, nil
}

func GetSettings() (*Settings, error) {
	var settings Settings
	result := db.First(&settings)

	if result.Error != nil {

		// Check if it's a 'record not found' error
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			// Creating default SlackSettings
			defaultSlackSettings := SlackSettings{
				Enabled:   false,
				ChannelID: "",
				Workspace: "",
			}

			jsonSlack, err := json.Marshal(defaultSlackSettings)
			if err != nil {
				return nil, err
			}

			defaultAuditLog := AuditLoggingSettings{Enabled: false}

			auditJSON, err := json.Marshal(defaultAuditLog)
			if err != nil {
				return nil, err
			}

			defaultSettings := Settings{
				SlackData:    string(jsonSlack),
				AuditLogData: string(auditJSON),
			}

			// Create the default settings in the database
			if err := db.Create(&defaultSettings).Error; err != nil {
				return nil, err
			}

			return &defaultSettings, nil
		}
		return nil, result.Error
	}

	// Deserialize SlackData into Slack struct
	// After unmarshalling
	if err := json.Unmarshal([]byte(settings.SlackData), &settings.Slack); err != nil {
		return nil, err
	}
	// Deserialize SlackData into Slack struct
	// After unmarshalling
	_ = json.Unmarshal([]byte(settings.AuditLogData), &settings.AuditLog)

	return &settings, nil
}

func UpdateSettings(updatedSettings *Settings) error {
	settingsDb, err := GetSettings()
	if err != nil {
		return err
	}

	// Serialize the updated SlackSettings to JSON
	updatedJson, err := json.Marshal(updatedSettings.Slack)
	if err != nil {
		return err
	}
	settingsDb.SlackData = string(updatedJson)

	// Serialize the updated SlackSettings to JSON
	auditlogUpdated, err := json.Marshal(updatedSettings.AuditLog)
	if err != nil {
		return err
	}
	settingsDb.AuditLogData = string(auditlogUpdated)

	// Update the existing record with new SlackData
	return db.Model(settingsDb).Update("SlackData", settingsDb.SlackData).Update("AuditLogData", settingsDb.AuditLogData).Error
}

func UpdateUser(user *UserData) error {
	return db.Model(user).Where("email = ?", user.Email).Update("title", user.Title).Error
}

func UpdateEvent(id uint, processed bool) error {
	result := db.Model(&EventQueue{}).Where("id = ?", id).Update("processed", processed)
	return result.Error // Return the error if there is one
}

func DeleteEvent(id uint) error {
	result := db.Model(&EventQueue{}).Where("id = ?", id).Delete(&EventQueue{})
	return result.Error // Return the error if there is one
}

func GetOpenEvents() (*[]EventQueue, error) {
	var events []EventQueue
	result := db.Where("processed = ?", false).Limit(10).Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return &events, nil
}

func GetAllEvents(limit int) (*[]EventQueue, error) {
	var eventQueues []EventQueue

	if limit <= 0 {
		limit = 50 // Default limit
	}

	result := db.Order("created_at desc").Limit(limit).Find(&eventQueues)
	if result.Error != nil {
		return nil, result.Error
	}
	return &eventQueues, nil
}

func SaveOrUpdateUserData(name string, email string, picture string) error {
	var existingUserData UserData

	// First, try to find the existing user data by email
	result := db.Where("email = ?", email).First(&existingUserData)

	// Handle the case where the user data might not exist
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If not found, create a new record
		newUserData := &UserData{
			Name:    name,
			Email:   email,
			Picture: picture,
		}
		return db.Create(newUserData).Error
	} else if result.Error != nil {
		// Handle other potential errors
		return result.Error
	}

	// If found, update the existing record
	existingUserData.Name = name
	existingUserData.Picture = picture
	return db.Save(&existingUserData).Error
}

func ChangeVulnerabilityStatus(id uint, status string) error {
	// Assuming `db` is your *gorm.DB instance

	// Update the status of the vulnerability
	result := db.Model(&JSONData{}).Where("id = ?", id).Update("Status", status)
	return result.Error
}

func GetUserDataByEmail(email string) (*UserData, error) {
	var userData UserData
	result := db.Where("email = ?", email).First(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

func GetAllUsers() (*[]UserData, error) {
	var userData []UserData
	result := db.Find(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

// createJSONData saves new JSON data to the database
func CreateJSONData(jsonData *JSONData) {
	db.Create(jsonData)
}

func SetVulnerabilitySlackUrl(id uint, url string) error {
	result := db.Model(&JSONData{}).Where("id = ?", id).Update("SlackUrl", url)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func UpdateVulnerability(jsonData *JSONData) error {
	// Assuming `db` is your *gorm.DB instance

	// Directly attempt to update the model using the primary key (ID)
	result := db.Model(&JSONData{}).Where("id = ?", jsonData.ID).Updates(jsonData)
	if result.Error != nil {
		// Handle the error, could be record not found or any other DB related error
		return result.Error
	}

	// RowsAffected can tell you how many records were updated
	if result.RowsAffected == 0 {
		// No records updated, which can indicate that the record wasn't found
		return errors.New("no records updated, record may not exist")
	}

	return nil // Return nil if no error occurred
}

func AllVulnerabilities() ([]JSONData, error) {
	var jsonData []JSONData
	result := db.Preload("Project").Order("created_at desc").Find(&jsonData) // Preload Project data
	return jsonData, result.Error
}

func CountOWASPCategories() (map[string]int, error) {
	var jsonData []JSONData
	result := db.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	categoryCounts := make(map[string]int)
	for _, data := range jsonData {
		var vuln Vulnerability
		// Assuming the vulnerability data is nested under a 'vulnerability' key
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			// Handle the error, perhaps continue to the next item
			continue
		}
		category := vuln.Category
		if category == "" {
			category = "uncategorized"
		}
		categoryCounts[category]++
	}

	return categoryCounts, nil
}

type VulnerabilityData struct {
	Category    string `json:"category"`
	Criticality string `json:"criticality"`
}

func FetchOWASPCriticalities() (map[string]map[string]int, error) {
	var jsonData []JSONData
	result := db.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	owaspData := make(map[string]map[string]int)
	for _, data := range jsonData {
		var vuln VulnerabilityData
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			return nil, err
		}

		// If category is empty, default it to "Uncategorized"
		category := vuln.Category
		if category == "" {
			category = "Uncategorized"
		}

		// If the category doesn't exist yet, initialize the criticality count maps
		if _, exists := owaspData[category]; !exists {
			owaspData[category] = map[string]int{
				"low":      0,
				"medium":   0,
				"high":     0,
				"critical": 0,
			}
		}

		// Increment the appropriate criticality count
		switch vuln.Criticality {
		case "low", "medium", "high", "critical":
			owaspData[category][vuln.Criticality]++
		}
	}

	return owaspData, nil
}

func CountUnresolvedTasks() (int, error) {
	var count int64
	result := db.Model(&JSONData{}).Where("status NOT IN (?)", []string{"Resolved", "Rejected"}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(count), nil
}

// CountCriticalities returns a map with the count of each criticality level
func CountCriticalities() (map[string]int, error) {
	var jsonData []JSONData
	result := db.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	criticalityCounts := make(map[string]int)
	for _, data := range jsonData {
		var vuln Vulnerability
		// Assuming the vulnerability data is nested under a 'vulnerability' key
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			// Handle the error, perhaps continue to the next item
			continue
		}

		criticalityCounts[vuln.Criticality]++
	}

	return criticalityCounts, nil
}

// getJSONData retrieves JSON data from the database
func GetJSONData(id uint) (JSONData, error) {
	var jsonData JSONData
	result := db.First(&jsonData, id)
	return jsonData, result.Error
}

func CountJSONData() (int64, error) {
	var count int64
	result := db.Model(&JSONData{}).Count(&count)
	return count, result.Error
}

func CountProjects() (int64, error) {
	var count int64
	result := db.Model(&ProjectData{}).Count(&count)
	return count, result.Error
}

func CountBugBounties() (int64, error) {
	var count int64
	var bugBountyProjectIDs []uint

	// Fetch IDs of projects with bug bounties
	err := db.Model(&ProjectData{}).Where("is_bug_bounty = ?", true).Pluck("id", &bugBountyProjectIDs).Error
	if err != nil {
		return 0, err
	}

	// Count vulnerabilities associated with those projects
	err = db.Model(&JSONData{}).Where("project_id IN (?)", bugBountyProjectIDs).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetProject(id uint) (ProjectData, error) {
	var project ProjectData
	result := db.First(&project, id)
	return project, result.Error
}

func GetProjects(query string) ([]ProjectData, error) {
	var projects []ProjectData

	// Prepare the database query
	db := db
	if query != "" {
		db = db.Where("project_name LIKE ?", "%"+query+"%")
	}

	// Sort the results by ProjectName in ascending order
	db = db.Order("project_name ASC").Find(&projects)

	if db.Error != nil {
		return nil, db.Error
	}

	return projects, nil
}

func CountProjectVulnerabilities(projectID uint) (int64, error) {
	var count int64
	err := db.Model(&JSONData{}).Where("project_id = ?", projectID).Count(&count).Error
	if err != nil {
		return 0, err // handle the error appropriately
	}

	return count, nil
}

func GetProjectVulnerabilities(projectID uint) ([]JSONData, error) {
	var jsonData []JSONData
	err := db.Where("project_id = ?", projectID).Order("created_at desc").Find(&jsonData).Error

	return jsonData, err
}

func deleteVulnerabilitiesAndImages(tx *gorm.DB, projectID *uint, ID *uint) error {
	var jsonData []JSONData
	appConfig, _ := config.LoadConfig()

	// Determine the query based on the provided IDs
	var err error
	if ID != nil {
		// Find specific jsonData entry by ID
		err = tx.Where("id = ?", *ID).Find(&jsonData).Error
	} else if projectID != nil {
		// Find all jsonData entries for a project
		err = tx.Where("project_id = ?", *projectID).Find(&jsonData).Error
	}

	if err != nil {
		return err
	}

	// Delete jsonData based on the provided IDs
	if ID != nil {
		// Delete specific jsonData entry by ID
		err = tx.Where("id = ?", *ID).Delete(&JSONData{}).Error
	} else {
		// Delete all jsonData entries for a project
		err = tx.Where("project_id = ?", *projectID).Delete(&JSONData{}).Error
	}

	if err != nil {
		return err
	}

	// Delete images
	for _, data := range jsonData {
		var vulnerability map[string]interface{}
		if err := json.Unmarshal(data.Vulnerability, &vulnerability); err != nil {
			return err
		}
		if images, ok := vulnerability["images"].([]interface{}); ok {
			for _, img := range images {
				if imageName, ok := img.(string); ok {
					imagePath := filepath.Join(appConfig.Database.Path, "/images", imageName)
					if err := os.Remove(imagePath); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func DeleteProjectAndAssets(projectID uint) error {
	// Start a transaction
	tx := db.Begin()

	if err := deleteVulnerabilitiesAndImages(tx, &projectID, nil); err != nil {
		tx.Rollback()
		return err
	}

	// Delete ProjectData
	if err := tx.Where("id = ?", projectID).Delete(&ProjectData{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func DeleteVulnerability(vulnerabilityID uint) error {
	// Start a transaction
	tx := db.Begin()

	if err := deleteVulnerabilitiesAndImages(tx, nil, &vulnerabilityID); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}
