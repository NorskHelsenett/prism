package database

import (
	"encoding/json"
	"errors"
	"os"

	"runtime"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"strings"

	"prism/config"
	"prism/models"
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
	Comments      datatypes.JSON
	Revisions     datatypes.JSON
}

// ImageData is a model for storing image metadata and binary data
type ImageData struct {
	gorm.Model
	Filename string
	Data     []byte
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
	Metrics      Metrics              `gorm:"-" json:"metrics"`
	MFAEnabled   bool                 `gorm:"default:false"`
}

type Metrics struct {
	Memory       float64
	DatabaseSize float64
}

type UserData struct {
	gorm.Model
	Email     string
	Name      string
	Picture   string
	Role      string `gorm:"default:visitor"`
	Title     string `gorm:"default:My title"`
	OTPSecret string `json:"-"`
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
	Error     string
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
	var err error
	db, err = gorm.Open(sqlite.Open(config.AppConfig.Database.Path+"/prism.db?cache=shared&_synchronous=FULL"), &gorm.Config{})
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
	db.AutoMigrate(&ImageData{})

	db.Exec(`
    CREATE TRIGGER IF NOT EXISTS jsondata_insert AFTER INSERT ON json_data
		BEGIN
    	INSERT INTO event_queues (table_id, table_name, created_at) VALUES (NEW.id, 'vulnerability', CURRENT_TIMESTAMP);
		END;
		`)
}

func CloseConnection() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err // return error if failed to get underlying database
	}

	return sqlDB.Close() // close the underlying SQL database
}

func RecordAuditLog(log AuditLog) error {
	return db.Create(&log).Error
}

func SetEventProcessed(event *EventQueue) {
	db.Model(&event).Update("processed", true).Update("error", event.Error)
}

func CleanUpDatabase() error {
	// Wrap the cleanup in a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// List of models to be hard-deleted
	models := []interface{}{&JSONData{}, &ProjectData{}, &UserData{}, &ImageData{}}

	for _, model := range models {
		// Perform hard delete on each model
		if err := tx.Unscoped().Where("deleted_at IS NOT NULL").Delete(model).Error; err != nil {
			tx.Rollback() // Rollback the transaction in case of error
			return err
		}
	}

	// Step 1: Collect all image references from JSONData
	var jsonDataList []JSONData
	if err := tx.Find(&jsonDataList).Error; err != nil {
		tx.Rollback()
		return err
	}

	imageRefs := make(map[string]bool)
	for _, jsonData := range jsonDataList {
		var vulnerability struct {
			Images []string `json:"images"`
		}

		if err := json.Unmarshal(jsonData.Vulnerability, &vulnerability); err != nil {
			continue // Handle error or log as needed
		}

		for _, imageRef := range vulnerability.Images {
			imageRefs[imageRef] = true
		}
	}

	// Step 2: Fetch all ImageData records
	var images []ImageData
	if err := tx.Find(&images).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Step 3: Mark images as deleted if not found in imageRefs
	for _, image := range images {
		if _, found := imageRefs[image.Filename]; !found {
			// Image not found in JSONData references, mark as deleted
			if err := tx.Model(&ImageData{}).Where("filename = ?", image.Filename).Update("deleted_at", time.Now()).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Commit the transaction for hard deletes
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// SQLite specific optimizations
	if err := optimizeSQLite(db); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return nil
}

func optimizeSQLite(db *gorm.DB) error {
	// List of SQL commands for maintenance and optimization
	commands := []string{
		"VACUUM;",  // Clean up the database, reduce file size
		"REINDEX;", // Rebuild all indices
		"ANALYZE;", // Analyze the database, gather statistics
		// Add more commands as necessary
	}

	for _, cmd := range commands {
		if err := db.Exec(cmd).Error; err != nil {
			return err
		}
	}

	return nil
}

func GetProjectIdFromVulnerabilityID(findingsID uint) (uint, error) {
	var projectID struct {
		ProjectID *uint
	}

	result := db.Model(&JSONData{}).Select("project_id").Where("id = ?", findingsID).First(&projectID)

	if result.Error != nil {
		return 0, result.Error
	}

	if projectID.ProjectID == nil {
		return 0, errors.New("project ID not found for the given vulnerability ID")
	}

	return *projectID.ProjectID, nil
}

func HasClientAccessToProject(email, projectID string) (bool, error) {
	var project ProjectData
	result := db.First(&project, projectID)

	if result.Error != nil {
		return false, result.Error
	}

	emails := strings.Split(project.ClientEmail, ",")
	for _, clientEmail := range emails {
		if strings.TrimSpace(clientEmail) == email {
			return true, nil
		}
	}

	emails = strings.Split(project.HackerName, ",")
	for _, hackerEmail := range emails {
		if strings.TrimSpace(hackerEmail) == email {
			return true, nil
		}
	}

	return false, nil
}

func CountAllAudits() (uint, error) {
	var count int64
	result := db.Model(&AuditLog{}).Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return uint(count), nil
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

func PersistOTPSecret(email string, secret string) error {
	return db.Model(&UserData{}).Where("email = ?", email).Update("otp_secret", secret).Error
}

func DeleteOTPCode(email string) error {
	return db.Model(&UserData{}).Where("email = ?", email).Update("otp_secret", nil).Error
}

func GetOTPCode(email string) (string, error) {
	var user UserData
	result := db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		return "", result.Error
	}

	return user.OTPSecret, nil
}

func CheckForOtpEnabled(email string) (bool, error) {
	var user UserData
	result := db.Where("email = ?", email).First(&user)

	// Check for database-related errors
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// No record found is not an error in this context
			return false, nil
		}
		// Some other error occurred
		return false, result.Error
	}

	// If OTPSecret is not empty, OTP is enabled
	return user.OTPSecret != "", nil
}

func GetSettings(calculateMetrics bool) (*Settings, error) {
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

	if calculateMetrics {
		// calculate database size in MB, RAM Size AND CPU Usage
		settings.Metrics.DatabaseSize, _ = getDatabaseSize()
		settings.Metrics.Memory = getMemoryUsage()
	}

	return &settings, nil
}

func getMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Sys)
}

func getDatabaseSize() (float64, error) {
	dbPath := config.AppConfig.Database.Path + "/prism.db"
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return 0, err
	}
	sizeInMB := float64(fileInfo.Size())
	return sizeInMB, nil
}

func UpdateSettings(updatedSettings *Settings) error {
	settingsDb, err := GetSettings(false)
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
	// @todo fix this db.Model(settingsDb) updates ALL fields, db.Model(&Settings{}) will only update updates fields, if that is what i want
	return db.Model(settingsDb).Update("SlackData", settingsDb.SlackData).Update("AuditLogData", settingsDb.AuditLogData).Update("mfa_enabled", updatedSettings.MFAEnabled).Error
}

func UpdateUser(user *UserData) error {
	return db.Model(&UserData{}).Where("email = ?", user.Email).Update("role", user.Role).Error
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

func DeleteComment(id uint, cid string) error {
	var jsonData JSONData
	db.Where("id = ?", id).First(&jsonData)

	var filteredComments []models.Comment
	var comments []models.Comment
	if jsonData.Comments != nil {
		if err := json.Unmarshal(jsonData.Comments, &comments); err != nil {
			return err
		}
	}
	for _, comment := range comments {
		if comment.ID != cid {
			filteredComments = append(filteredComments, comment)
		}
	}

	commentsJSON, err := json.Marshal(filteredComments)
	if err != nil {
		return err
	}
	jsonData.Comments = datatypes.JSON(commentsJSON)

	result := db.Model(&JSONData{}).Where("id = ?", id).Update("Comments", jsonData.Comments)
	return result.Error
}

func UpdateComment(id uint, comment models.Comment) error {
	var jsonData JSONData
	db.Where("id = ?", id).First(&jsonData)

	var filteredComments []models.Comment
	var comments []models.Comment
	if jsonData.Comments != nil {
		if err := json.Unmarshal(jsonData.Comments, &comments); err != nil {
			return err
		}
	}
	for _, oldComment := range comments {
		if oldComment.ID != comment.ID {
			filteredComments = append(filteredComments, oldComment)
		} else {
			comment.ParentID = oldComment.ParentID
			filteredComments = append(filteredComments, comment)
		}
	}

	commentsJSON, err := json.Marshal(filteredComments)
	if err != nil {
		return err
	}
	jsonData.Comments = datatypes.JSON(commentsJSON)

	result := db.Model(&JSONData{}).Where("id = ?", id).Update("Comments", jsonData.Comments)
	return result.Error
}

func PersistComment(id uint, comment models.Comment) error {

	var jsonData JSONData
	db.Where("id = ?", id).First(&jsonData)

	var comments []models.Comment
	if jsonData.Comments != nil {
		if err := json.Unmarshal(jsonData.Comments, &comments); err != nil {
			return err
		}
	}

	comments = append(comments, comment)
	commentsJSON, err := json.Marshal(comments)
	if err != nil {
		return err
	}
	jsonData.Comments = datatypes.JSON(commentsJSON)

	result := db.Save(jsonData)
	// result := db.Model(&JSONData{}).Where("id = ?", id).Update("Comments", jsonData.Comments)
	return result.Error
}

func ChangeVulnerabilityStatus(id uint, status, email string) error {

	var jsonData JSONData

	db.Where("id = ?", id).First(&jsonData).Select("Status", "Revisions")

	var revision = models.Revision{
		UserEmail: email,
		CreatedAt: time.Now(),
		OldValue:  jsonData.Status,
		NewValue:  status,
		Property:  "status",
	}

	var revisions []models.Revision
	if jsonData.Revisions != nil {
		if err := json.Unmarshal(jsonData.Revisions, &revisions); err != nil {
			return err
		}
	}
	revisions = append(revisions, revision)

	// Assuming j.RevisionsTemp is a []models.Revision
	revisionsJSON, err := json.Marshal(revisions)
	if err != nil {
		return err
	}

	result := db.Model(&JSONData{}).Where("id = ?", id).Update("Status", status).Update("Revisions", revisionsJSON)
	return result.Error
}

func GetUserDataByEmail(email string) (*UserData, error) {
	var userData UserData

	result := db.Select("Name", "Picture", "Email").Where("email = ?", email).First(&userData)

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

func AllVulnerabilities(all bool, email string) ([]JSONData, error) {
	var jsonData []JSONData

	if all {
		// Admin: Get all vulnerabilities
		result := db.Preload("Project").Order("json_data.created_at desc").Find(&jsonData)
		return jsonData, result.Error
	} else {
		// Non-admin: Find all project IDs where the email is in "client_email"
		var projectIDs []uint
		emailPattern := "%" + email + "%"
		db.Model(&ProjectData{}).Where("client_email LIKE ?", emailPattern).Or("hacker_name LIKE ?", emailPattern).Pluck("id", &projectIDs)

		if len(projectIDs) == 0 {
			// No projects found for this email, return empty jsonData
			return jsonData, nil
		}

		// Get vulnerabilities for those projects
		result := db.Preload("Project").
			Where("project_id IN ?", projectIDs).
			Or("found_by LIKE ?", email).
			Order("json_data.created_at desc").
			Find(&jsonData)
		return jsonData, result.Error
	}
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
				"information": 0,
				"low":         0,
				"medium":      0,
				"high":        0,
				"critical":    0,
			}
		}

		// Increment the appropriate criticality count
		switch vuln.Criticality {
		case "information", "low", "medium", "high", "critical":
			owaspData[category][vuln.Criticality]++
		}
	}

	return owaspData, nil
}

func CountByStatus() (map[string]int, error) {
	var results []struct {
		Status string
		Count  int
	}
	statusCounts := make(map[string]int)

	// Group the results by 'Status' and count each group
	if err := db.Model(&JSONData{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// Fill the map with the status counts
	for _, result := range results {
		statusCounts[result.Status] = result.Count
	}

	return statusCounts, nil
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

func GetProjectsFor(email string) ([]ProjectData, error) {
	var projects []ProjectData

	// Prepare the database query to find projects where the email is in the ClientEmail list
	emailPattern := "%" + email + "%"
	db := db.Where("client_email LIKE ?", emailPattern).Or("hacker_name LIKE ?", emailPattern)

	// Execute the query and sort the results by ProjectName in ascending order
	db = db.Order("project_name ASC").Find(&projects)

	if db.Error != nil {
		return nil, db.Error
	}
	return projects, nil
}

func GetProjects() ([]ProjectData, error) {
	var projects []ProjectData

	// Sort the results by ProjectName in ascending order
	result := db.Order("project_name ASC").Find(&projects)

	if result.Error != nil {
		return nil, result.Error
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

func SaveImage(filename string, data []byte) error {
	// Create an ImageData instance
	img := ImageData{Filename: filename, Data: data}
	// Save to database
	return db.Create(&img).Error
}

func DeleteImage(filename string) error {
	// Delete the record
	return db.Where("filename = ?", filename).Delete(&ImageData{}).Error
}

func GetImage(filename string) (*ImageData, error) {
	var img ImageData
	err := db.Where("filename = ?", filename).First(&img).Error
	return &img, err
}

func deleteVulnerabilitiesAndImages(tx *gorm.DB, projectID *uint, ID *uint) error {
	var jsonData []JSONData

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
					// Delete the image using the database package method
					tx.Where("filename = ?", imageName).Delete(&ImageData{})
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
