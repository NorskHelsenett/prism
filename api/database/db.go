package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

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

type Subscriber struct {
	gorm.Model
	Email        string
	Subscription datatypes.JSON
}

func CreateProject(project *ProjectData) error {
	result := db.Create(project) // Create a new record
	return result.Error
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

type AssessmentJSON struct {
	gorm.Model
	Assessment datatypes.JSON
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
	SlackData           string               `json:"-" gorm:"column:slack_settings"`
	Slack               SlackSettings        `gorm:"-" json:"slack"`
	AuditLogData        string               `json:"-" gorm:"column:auditlog_settings"`
	AuditLog            AuditLoggingSettings `gorm:"-" json:"auditlog"`
	Metrics             Metrics              `gorm:"-" json:"metrics"`
	MFAEnabled          bool                 `gorm:"default:false"`
	WebPushNotification string               `gorm:"column:vaapi_keys" json:"-"`
}

type WebPushNotificationSettings struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type Metrics struct {
	Memory       float64
	DatabaseSize float64
}

type UserData struct {
	gorm.Model
	Email         string         `json:"email" gorm:"uniqueIndex"`
	Name          string         `json:"name"`
	Picture       string         `json:"picture"`
	Role          string         `json:"role" gorm:"default:visitor"`
	Title         string         `json:"title" gorm:"default:My title"`
	Active        *bool          `json:"active" gorm:"default:true"`
	OTPSecret     string         `json:"-"`
	Notifications datatypes.JSON `json:"-"`
	Settings      datatypes.JSON `json:"-"`
}

type Vulnerability struct {
	Criticality string `json:"criticality"`
	Category    string `json:"category"`
	ProjectID   uint   `json:"projectID"`
}

type WebAuthnCredential struct {
	gorm.Model
	Email          string `gorm:"index"`
	Name           string
	CredentialData []byte
}

type EventQueue struct {
	ID        uint `gorm:"primaryKey"`
	TableID   uint
	TableName string
	Error     string
	Processed bool      `gorm:"default:false;index:idx_processed"`
	CreatedAt time.Time `gorm:"index:idx_created_at,autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
	Kind      models.EventKind
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
	dbPath := config.AppConfig.Database.Path
	fullPath := filepath.Join(dbPath, "prism.db")

	// Ensure the directory exists
	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory '%s': %v", dbPath, err)
	}

	var err error
	db, err = gorm.Open(sqlite.Open(fullPath+"?cache=shared&_synchronous=FULL"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database at '%s': %v", config.AppConfig.Database.Path, err)
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
	db.AutoMigrate(&AssessmentJSON{})
	db.AutoMigrate(&Subscriber{})
	db.AutoMigrate(&models.APIKey{})
	db.AutoMigrate(&models.Report{})
	db.AutoMigrate(&models.ReportVersion{})
	db.AutoMigrate(&models.SharedDocument{})
	db.AutoMigrate(&models.Team{})
	db.AutoMigrate(&WebAuthnCredential{})
	db.AutoMigrate(&models.Note{})

	initNotesFTS()

	// Create the accessible_vulnerabilities view
	initAccessibleVulnerabilitiesView()

	db.Exec("DROP TRIGGER IF EXISTS jsondata_insert;")
	db.Exec("DROP TRIGGER IF EXISTS jsondata_update_comments;")

	db.Exec(`
    CREATE TRIGGER IF NOT EXISTS jsondata_insert AFTER INSERT ON json_data
		BEGIN
    	INSERT INTO event_queues (table_id, table_name, created_at, kind) VALUES (NEW.id, 'vulnerability', CURRENT_TIMESTAMP, 1);
		END;
		`)

	db.Exec(`
		CREATE TRIGGER IF NOT EXISTS jsondata_update_comments
		AFTER UPDATE OF comments ON json_data
		FOR EACH ROW
		WHEN (OLD.comments IS NOT NEW.comments)
		BEGIN
				INSERT INTO event_queues (table_id, table_name, created_at, kind)
				VALUES (
						NEW.id,
						'vulnerability',
						CURRENT_TIMESTAMP,
						2
				);
		END;
	`)
}

func initAccessibleVulnerabilitiesView() {
	// Drop the view if it exists
	if err := db.Exec("DROP VIEW IF EXISTS accessible_vulnerabilities;").Error; err != nil {
		log.Fatalf("Failed to drop view accessible_vulnerabilities: %v", err)
	}

	// Create the view
	viewSQL := `
		CREATE VIEW accessible_vulnerabilities AS
		SELECT
				json_data.id AS id,
				json_extract(json_data.vulnerability, '$.visibility') AS visibility,
				json_extract(json_data.vulnerability, '$.assignedTo') AS assigned_to,
				json_data.found_by,
				json_data.project_id,
				project_data.client_email,
				project_data.hacker_name
		FROM json_data
		LEFT JOIN project_data ON json_data.project_id = project_data.id;
	`

	if err := db.Exec(viewSQL).Error; err != nil {
		log.Fatalf("Failed to create view accessible_vulnerabilities: %v", err)
	}
}

func CloseConnection() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err // return error if failed to get underlying database
	}

	return sqlDB.Close() // close the underlying SQL database
}

func GetAllTeams() ([]models.Team, error) {
	var teams []models.Team
	result := db.Find(&teams)
	return teams, result.Error
}

func GetTeamByID(id uint) (*models.Team, error) {
	var team models.Team
	result := db.First(&team, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &team, nil
}

func CreateTeam(team *models.Team) error {
	return db.Create(team).Error
}

func UpdateTeam(team *models.Team) error {
	return db.Save(team).Error
}

func DeleteTeam(id uint) error {
	return db.Delete(&models.Team{}, id).Error
}

func ArchiveTeam(id uint) error {
	return db.Model(&models.Team{}).Where("id = ?", id).Update("archived", true).Error
}

func AddMemberToTeam(teamID uint, email string) error {
	var team models.Team
	if err := db.First(&team, teamID).Error; err != nil {
		return err
	}

	for _, member := range team.MembersJSON {
		if member == email {
			return nil // Member already exists
		}
	}

	team.MembersJSON = append(team.MembersJSON, email)
	return db.Save(&team).Error
}

func RemoveMemberFromTeam(teamID uint, email string) error {
	var team models.Team
	if err := db.First(&team, teamID).Error; err != nil {
		return err
	}

	for i, member := range team.MembersJSON {
		if member == email {
			team.MembersJSON = append(team.MembersJSON[:i], team.MembersJSON[i+1:]...)
			return db.Save(&team).Error
		}
	}

	return nil // Member not found
}

func GetTeamsByMemberEmail(email string) ([]models.Team, error) {
	var teams []models.Team
	result := db.Where("archived = ?", false).Where("JSON_CONTAINS(members, ?)", fmt.Sprintf("\"%s\"", email)).Find(&teams)
	return teams, result.Error
}

type ProfileResponse struct {
	Teams []models.Team `json:"teams"`
	Users []UserData    `json:"users"`
}

func GetAllProfilesWithTeams() (*ProfileResponse, error) {
	var teams []models.Team
	var users []UserData

	if err := db.Where("archived = ?", false).Find(&teams).Error; err != nil {
		return nil, err
	}

	if err := db.Select("Name", "Email", "Active").Find(&users).Error; err != nil {
		return nil, err
	}

	inactive := make(map[string]bool)
	for _, u := range users {
		if u.Active != nil && !*u.Active {
			inactive[u.Email] = true
		}
	}

	// Filter inactive members from team member lists
	for i := range teams {
		filtered := make([]string, 0, len(teams[i].MembersJSON))
		for _, email := range teams[i].MembersJSON {
			if !inactive[email] {
				filtered = append(filtered, email)
			}
		}
		teams[i].MembersJSON = filtered
	}

	usersInTeams := make(map[string]bool)
	for _, team := range teams {
		for _, email := range team.MembersJSON {
			usersInTeams[email] = true
		}
	}

	var individualUsers []UserData
	for _, user := range users {
		if inactive[user.Email] {
			continue
		}
		if !usersInTeams[user.Email] {
		}
		individualUsers = append(individualUsers, user)
	}

	return &ProfileResponse{
		Teams: teams,
		Users: individualUsers,
	}, nil
}

func DeleteNotifications(email string) error {
	var userData UserData

	// Fetch the user data by email
	result := db.Where("email = ?", email).First(&userData)
	if result.Error != nil {
		return result.Error
	}

	userData.Notifications = nil

	result = db.Save(&userData)
	return result.Error
}

func GetAllSubscribers() ([]Subscriber, error) {
	var subscriberList []Subscriber
	result := db.Find(&subscriberList)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriberList, nil
}

func ExistsShare(token string) (*models.SharedDocument, error) {
	var share models.SharedDocument

	query := db.Table("shared_documents").
		Select("shared_documents.*").
		Joins("INNER JOIN json_data ON shared_documents.document_id = json_data.id").
		Where("json_data.deleted_at IS NULL").
		Where("shared_documents.share_token = ?", token).
		First(&share)

	if query.Error != nil {
		return nil, query.Error
	}

	return &share, nil
}

func GetAllShares(email string) (*[]models.SharedDocument, error) {
	var shares []models.SharedDocument
	result := db.Where("shared_by_email = ?", email).Find(&shares)

	if result.Error != nil {
		return nil, result.Error
	}
	return &shares, nil
}

func GetShareDocument(id uint) (*models.SharedDocument, error) {
	var share models.SharedDocument
	result := db.Where("document_id = ?", id).First(&share)

	if result.Error != nil {
		return nil, result.Error
	}
	return &share, nil
}

func DeleteShareDocument(id uint) error {
	var shareDoc models.SharedDocument
	shareDoc.DocumentID = id
	if err := db.Delete(&shareDoc).Error; err != nil {
		return err
	}
	return nil
}

func PersistShareDocument(share *models.SharedDocument) error {
	if err := db.Save(&share).Error; err != nil {
		return err
	}
	return nil
}

func AppendSubscriber(email string, subs []byte) error {
	subscriber := Subscriber{
		Email:        email,
		Subscription: subs,
	}

	if err := db.Create(&subscriber).Error; err != nil {
		return err
	}

	return nil
}

func CreateNotification(email string, notification models.Notification) error {
	var userData UserData

	// Fetch the user data by email
	result := db.Where("email = ?", email).First(&userData)
	if result.Error != nil {
		// Check if the error is a RecordNotFound error
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Fail softly by returning nil
			return nil
		}
		// For other types of errors, return the error
		return result.Error
	}

	// Unmarshal the JSON notifications into a slice
	var notifications []models.Notification
	// Check if the JSON field has content before unmarshalling
	if len(userData.Notifications) > 0 {
		if err := json.Unmarshal(userData.Notifications, &notifications); err != nil {
			return err
		}
	} else {
		// If the JSON is nil or empty, initialize it to an empty slice to avoid null in JSON output
		notifications = []models.Notification{}
	}

	notifications = append(notifications, notification)

	// Marshal the updated notifications slice back into JSON
	updatedNotifications, err := json.Marshal(notifications)
	if err != nil {
		return err
	}
	userData.Notifications = updatedNotifications

	// Save the updated user data back to the database
	result = db.Save(&userData)
	return result.Error
}

func MarkNotificationAsRead(email string, notificationTime time.Time) error {
	var userData UserData

	// Fetch the user data by email
	result := db.Where("email = ?", email).First(&userData)
	if result.Error != nil {
		return result.Error
	}

	// Unmarshal the JSON notifications into a slice
	var notifications []models.Notification
	if err := json.Unmarshal(userData.Notifications, &notifications); err != nil {
		return err
	}

	// Mark the notification as read
	updated := false
	for i, notification := range notifications {
		if notification.When.Equal(notificationTime) && !notification.IsRead {
			notifications[i].IsRead = true
			updated = true
			break
		}
	}

	// If the notification was found and updated, marshal the slice back to JSON and update the record
	if updated {
		updatedNotifications, err := json.Marshal(notifications)
		if err != nil {
			return err
		}
		userData.Notifications = updatedNotifications
		result = db.Save(&userData)
		return result.Error
	}

	// If no updates were made, return nil to indicate no error occurred
	return nil
}

func GetNotifications(email string) ([]models.Notification, error) {
	var userData UserData
	result := db.Where("email = ?", email).First(&userData)
	if result.Error != nil {
		return nil, result.Error
	}

	var notifications []models.Notification

	// Check if the JSON field has content before unmarshalling
	if len(userData.Notifications) > 0 {
		if err := json.Unmarshal(userData.Notifications, &notifications); err != nil {
			return nil, err
		}
	} else {
		// If the JSON is nil or empty, initialize it to an empty slice to avoid null in JSON output
		notifications = []models.Notification{}
	}

	return notifications, nil
}

func UpdateAssassment(assessment models.Assessment, id uint) error {
	data, err := json.Marshal(assessment)
	if err != nil {
		return err
	}

	return db.Model(&AssessmentJSON{}).Where("id = ?", id).Update("assessment", data).Error
}

func ValidateAPIKey(hashedAPIKey string) (string, bool) {
	var storedKey models.APIKey

	err := db.Where("hashed_api_key = ? AND expiry_at > ?", hashedAPIKey, time.Now()).First(&storedKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// API key is invalid or expired
			return "", false
		}
		// Other errors while accessing the database
		return "", false
	}

	// The API key is valid, update the updated_at column
	if err := db.Model(&storedKey).Update("updated_at", time.Now()).Error; err != nil {
		// Handle error in updating the record
		return "", false
	}

	email := storedKey.Email

	return email, true
}

var ErrForbidden = errors.New("forbidden")

func DeleteApiKey(id uint, email string) error {
	var apikey models.APIKey
	result := db.Where("id = ?", id).First(&apikey)

	// Check for errors in retrieving the API key
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil // No record found, nothing to delete
		}
		return result.Error
	}

	// Check if the email matches the API key's email
	if apikey.Email != email {
		return ErrForbidden // Email does not match, return 403 forbidden error
	}

	// Email matches, proceed to delete the API key
	if err := db.Delete(&apikey).Error; err != nil {
		return err
	}

	return nil
}

func GetApiKeys(email string) (*[]models.APIKey, error) {
	var apikeys []models.APIKey
	result := db.Where("email = ?", email).Find(&apikeys)

	if result.Error != nil {
		return nil, result.Error
	}
	return &apikeys, nil
}

func PersistApiKey(apikey *models.APIKey) (*models.APIKey, error) {
	if err := db.Create(&apikey).Error; err != nil {
		return nil, err
	}
	return apikey, nil
}

func PersistAssessment(assessment models.Assessment) (uint, error) {
	var assessmentJSON AssessmentJSON
	data, err := json.Marshal(assessment)
	if err != nil {
		return 0, err
	}

	assessmentJSON.Assessment = data
	if err := db.Create(&assessmentJSON).Error; err != nil {
		return 0, err
	}

	return assessmentJSON.ID, nil // Return the new ID
}

func PopulateProjectName(projectID uint) (string, error) {
	var project ProjectData
	result := db.First(&project, projectID).Select("project_name")
	if result.Error != nil {
		return "", result.Error
	}
	return project.ProjectName, nil
}

func DeleteAssessment(id uint) error {
	return db.Where("id = ?", id).Delete(&AssessmentJSON{}).Error
}

func RetrieveAssessment(id uint) (*AssessmentJSON, error) {
	var assessment AssessmentJSON
	result := db.First(&assessment, id)

	if result.Error != nil {
		return nil, result.Error
	}
	return &assessment, nil
}

func RetrieveAssessments(startDate, endDate string, page, pageSize int) ([]AssessmentJSON, error) {
	var assessments []AssessmentJSON

	// Calculate offset for pagination
	offset := (page - 1) * pageSize

	// Build a query to find assessments that overlap with the date range
	// An assessment overlaps if:
	// - Its end date is on or after the start of the range, AND
	// - Its start date is on or before the end of the range
	query := "json_extract(assessment, '$.dateTo') >= ? AND json_extract(assessment, '$.dateFrom') <= ?"

	// Execute the query with pagination
	result := db.Model(&AssessmentJSON{}).
		Where(query, startDate, endDate).
		Offset(offset).Limit(pageSize).
		Order("json_extract(assessment, '$.dateFrom') ASC").
		Find(&assessments)

	if result.Error != nil {
		return nil, result.Error
	}

	return assessments, nil
}

func RecordAuditLog(log AuditLog) error {
	return db.Create(&log).Error
}

func SetEventProcessed(event *EventQueue) {
	db.Model(&event).Update("processed", true).Update("error", event.Error)
}

func GetFilteredVulnerabilities(isGlobal bool, email string, year string, status string, severity string, projectID uint) ([]JSONData, error) {
	var jsonData []JSONData
	query := db.Preload("Project").
		Select(`
			json_data.*, 
			json_extract(json_data.vulnerability, '$.title') as vulnerability_title, 
			json_extract(json_data.vulnerability, '$.isPublicFacing') as vulnerability_isPublicFacing, 
			json_extract(json_data.vulnerability, '$.criticality') as vulnerability_criticality, 
			json_extract(json_data.vulnerability, '$.date') as vulnerability_date
		`)

	if !isGlobal {
		// Non-admin: Find all project IDs where the email is in "client_email" or "hacker_name"
		// Use exact match with comma separators to avoid substring attacks
		subQuery := db.Table("project_data").
			Select("id").
			Where("(',' || client_email || ',') LIKE ? OR (',' || hacker_name || ',') LIKE ?", "%,"+email+",%", "%,"+email+",%")

		query = query.Where(
			db.Where("project_id IN (?) AND deleted_at IS NULL", subQuery).
				Or("found_by = ? AND deleted_at IS NULL", email))
	}

	if year != "" {
		query = query.Where("strftime('%Y', json_extract(json_data.vulnerability, '$.date')) = ?", year)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("json_extract(json_data.vulnerability, '$.criticality') = ?", severity)
	}
	if projectID != 0 {
		query = query.Where("project_id = ?", projectID)
	}

	query = query.Order("json_extract(json_data.vulnerability, '$.date') desc")
	result := query.Find(&jsonData)
	return jsonData, result.Error
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

// Uses the accessible_vulnerabilities view for efficient DB-level filtering.
// globalViewer: true if the user's role has global view rights
func CanAccessVulnerability(vulnID uint, email string, role string, globalViewer bool) (bool, error) {
	if role == "admin" {
		return true, nil
	}
	var count int64
	query := db.Table("accessible_vulnerabilities").Where("id = ?", vulnID)
	if globalViewer {
		query = query.Where(`
            project_id IS NULL
            OR visibility IN ('published', 'public')
            OR assigned_to = ?
            OR found_by = ?
            OR ',' || COALESCE(client_email, '') || ',' LIKE ?
            OR ',' || COALESCE(hacker_name, '') || ',' LIKE ?
        `, email, email, "%,"+email+",%", "%,"+email+",%")
	} else {
		query = query.Where(`
            assigned_to = ?
            OR found_by = ?
            OR ',' || COALESCE(client_email, '') || ',' LIKE ?
            OR ',' || COALESCE(hacker_name, '') || ',' LIKE ?
        `, email, email, "%,"+email+",%", "%,"+email+",%")
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func HasAccessToProject(email string, projectID string, role string) (bool, error) {
	if role == "admin" {
		return true, nil
	}
	projectIDInt, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid project ID: %v", err)
	}
	var count int64
	// Direct query to project_data for efficiency
	query := db.Table("project_data").
		Where("id = ? AND deleted_at IS NULL", projectIDInt).
		Where("',' || COALESCE(client_email, '') || ',' LIKE ? OR ',' || COALESCE(hacker_name, '') || ',' LIKE ?", "%,"+email+",%", "%,"+email+",%")
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasWriteOnAllProjects reports whether the user has write access to *every*
// project in projectIDs. Used to gate Report create/edit/publish/delete:
// the author must have write on each linked project, mirroring the existing
// HasAccessToProject email-match rule (client_email / hacker_name).
//
// Admin and any role with "*" on /project bypasses the check.
func HasWriteOnAllProjects(email string, role string, projectIDs []uint, globalProject bool) (bool, error) {
	if role == "admin" || globalProject {
		return true, nil
	}
	if len(projectIDs) == 0 {
		return false, nil
	}
	for _, pid := range projectIDs {
		ok, err := HasAccessToProject(email, strconv.FormatUint(uint64(pid), 10), role)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// HasReadOnAnyProject reports whether the user can read at least one of the
// listed projects. Used to gate Report internal read: anyone with read on
// any linked project sees published versions.
func HasReadOnAnyProject(email string, role string, projectIDs []uint, globalProject bool) (bool, error) {
	if role == "admin" || globalProject {
		return true, nil
	}
	for _, pid := range projectIDs {
		ok, err := HasAccessToProject(email, strconv.FormatUint(uint64(pid), 10), role)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
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
		if strings.TrimSpace(strings.ToLower(clientEmail)) == strings.ToLower(email) {
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

func DeleteUser(id string) error {
	return db.Where("id = ?", id).Delete(&UserData{}).Error
}

func ToggleUserActive(id string) (*UserData, error) {
	var user UserData
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	newActive := !*user.Active
	user.Active = &newActive
	if err := db.Model(&user).Update("active", newActive).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetInactiveUserEmails() (map[string]bool, error) {
	var users []UserData
	if err := db.Where("active = ?", false).Select("email").Find(&users).Error; err != nil {
		return nil, err
	}
	emails := make(map[string]bool, len(users))
	for _, u := range users {
		emails[u.Email] = true
	}
	return emails, nil
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

func GetUserByEmail(email string) (*UserData, error) {
	var user UserData
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func SaveWebAuthnCredential(email, name string, credentialData []byte) error {
	cred := WebAuthnCredential{
		Email:          email,
		Name:           name,
		CredentialData: credentialData,
	}
	return db.Create(&cred).Error
}

func GetWebAuthnCredentials(email string) ([]WebAuthnCredential, error) {
	var creds []WebAuthnCredential
	result := db.Where("email = ?", email).Find(&creds)
	if result.Error != nil {
		return nil, result.Error
	}
	return creds, nil
}

func DeleteWebAuthnCredential(email, id string) error {
	return db.Where("email = ? AND id = ?", email, id).Delete(&WebAuthnCredential{}).Error
}

func DeleteAllWebAuthnCredentials(email string) error {
	return db.Where("email = ?", email).Delete(&WebAuthnCredential{}).Error
}

func HasWebAuthnCredentials(email string) (bool, error) {
	var count int64
	result := db.Model(&WebAuthnCredential{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func CountWebAuthnCredentials(email string) (int64, error) {
	var count int64
	result := db.Model(&WebAuthnCredential{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
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

func GetVAAPIprivateKey() (string, error) {
	settings, err := GetSettings(false)
	if err != nil {
		return "", err
	}

	var webpushSettings WebPushNotificationSettings

	err = json.Unmarshal([]byte(settings.WebPushNotification), &webpushSettings)
	if err != nil {
		return "", err
	}

	return webpushSettings.PrivateKey, nil
}

func ResetNotifications() error {
	// Retrieve settings without cache
	settings, err := GetSettings(false)
	if err != nil {
		return err
	}

	// Begin a transaction to ensure that deleting subscribers and updating settings are atomic
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Attempt to delete all subscribers from the database
	if err := tx.Exec("DELETE FROM subscribers").Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		return err
	}

	// Reset the WebPushNotification key in the settings
	settings.WebPushNotification = ""

	// Update the settings in the database
	if err := tx.Model(&settings).Update("WebPushNotification", settings.WebPushNotification).Error; err != nil {
		tx.Rollback() // Rollback the transaction on error
		return err
	}

	// Commit the transaction if all went well
	return tx.Commit().Error
}

func GetVAAPIpublicKey() (string, error) {
	settings, err := GetSettings(false)
	if err != nil {
		return "", err
	}

	var webpushSettings WebPushNotificationSettings

	if settings.WebPushNotification == "" {
		return "", nil
	}

	err = json.Unmarshal([]byte(settings.WebPushNotification), &webpushSettings)
	if err != nil {
		return "", err
	}

	return webpushSettings.PublicKey, nil
}

func CreateVAAPIprivateKey(privateKey, publicKey string) error {
	settings, err := GetSettings(false)
	if err != nil {
		return err
	}

	var webpushSettings WebPushNotificationSettings

	if settings.WebPushNotification != "" {
		err = json.Unmarshal([]byte(settings.WebPushNotification), &webpushSettings)
		if err != nil {
			return err
		}
		if webpushSettings.PrivateKey != "" {
			return errors.New("private key already exists")
		}
	}

	webpushSettings.PrivateKey = privateKey
	webpushSettings.PublicKey = publicKey

	webpushJson, err := json.Marshal(webpushSettings)
	if err != nil {
		return err
	}

	settings.WebPushNotification = string(webpushJson)

	return db.Model(settings).Update("WebPushNotification", settings.WebPushNotification).Error
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

func UpdateUserRole(user *UserData) error {
	return db.Model(&UserData{}).Where("email = ?", user.Email).Update("role", user.Role).Error
}

// GetPreferencesForUser retrieves user preferences by email
func GetPreferencesForUser(email string) (models.UserSettings, error) {
	var user UserData
	var preferences models.UserSettings

	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return preferences, result.Error
	}

	// If user has no settings yet, return empty preferences
	if len(user.Settings) == 0 {
		return preferences, nil
	}

	// Unmarshal the stored JSON settings
	if err := json.Unmarshal(user.Settings, &preferences); err != nil {
		return preferences, err
	}

	return preferences, nil
}

// PatchSettingsForUser updates user preferences
func PatchSettingsForUser(email string, preferences models.UserSettings) error {
	var user UserData

	// First find the user
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return result.Error
	}

	// Marshal the preferences to JSON
	settingsJSON, err := json.Marshal(preferences)
	if err != nil {
		return err
	}

	// Update the user's settings
	user.Settings = settingsJSON
	result = db.Save(&user)

	return result.Error
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
	if picture != "" {
		existingUserData.Picture = picture
	}
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
func CreateJSONData(jsonData *JSONData) error {
	// Check for existing vulnerability with same data
	var existingCount int64
	err := db.Model(&JSONData{}).
		Where("vulnerability = ?", jsonData.Vulnerability).
		Count(&existingCount).Error

	if err != nil {
		return err
	}

	if existingCount > 0 {
		return errors.New("this vulnerability has already been reported")
	}

	// Create the entry if no duplicate found
	return db.Create(jsonData).Error
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

// UpdateVulnerabilityJSON updates only the vulnerability JSON column for a given record ID.
// This is useful for PATCH-style partial updates where only parts of the JSON blob are changed.
func UpdateVulnerabilityJSON(id uint, vuln datatypes.JSON) error {
	result := db.Model(&JSONData{}).Where("id = ?", id).Update("vulnerability", vuln)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated, record may not exist")
	}

	return nil
}

// UpdateJSONDataFields updates only provided top-level fields of JSONData (metadata), leaving other fields untouched.
// fields is a map of column names to values to update.
func UpdateJSONDataFields(id uint, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	result := db.Model(&JSONData{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated, record may not exist")
	}

	return nil
}

type MinimalJSONData struct {
	ID            uint                  `json:"ID"`
	Vulnerability MinifiedVulnerability `json:"Vulnerability"`
	FoundBy       string                `json:"FoundBy"`
	ProjectID     uint                  `json:"ProjectID"`
	Project       struct {
		ID          uint   `json:"ID"`
		ProjectName string `json:"ProjectName"`
	} `json:"Project"`
	Status string `json:"Status"`
}

type MinifiedVulnerability struct {
	Title          string `json:"title"`
	IsPublicFacing bool   `json:"isPublicFacing"`
	Criticality    string `json:"criticality"`
	Date           string `json:"date"`
	Visibility     string `json:"visibility"`
	Category       string `json:"category"`
}

type VulnerabilitySearchParams struct {
	Query        string
	Criticality  []string
	Status       []string
	Location     string // "public", "internal", or ""
	HideClosed   bool
	AssignedToMe bool
	Year         string // e.g. "2026", or "" for all years
}

type vulnerabilityAccessEnvelope struct {
	Visibility string `json:"visibility"`
	AssignedTo string `json:"assignedTo"`
}

func isRestrictedVisibility(visibility string) bool {
	vis := strings.ToLower(strings.TrimSpace(visibility))
	switch vis {
	case "", "published", "public":
		return false
	case "undisclosed", "hidden", "private":
		return true
	default:
		return vis != ""
	}
}

// CanViewVulnerability determines whether a user has access to the provided vulnerability entry.
// Administrators (isGlobal == true) always have access. For restricted visibilities, only the reporter
// (FoundBy) and explicitly assigned user(s) may view the record.
func CanViewVulnerability(vulnerability JSONData, email string, isGlobal bool) bool {
	if isGlobal {
		return true
	}

	if strings.EqualFold(strings.TrimSpace(vulnerability.FoundBy), strings.TrimSpace(email)) {
		return true
	}

	var envelope vulnerabilityAccessEnvelope
	if err := json.Unmarshal(vulnerability.Vulnerability, &envelope); err != nil {
		// Fail-secure: deny access if we cannot parse the payload
		return false
	}

	if !isRestrictedVisibility(envelope.Visibility) {
		return true
	}

	for _, assigned := range strings.Split(envelope.AssignedTo, ",") {
		if strings.EqualFold(strings.TrimSpace(assigned), strings.TrimSpace(email)) {
			return true
		}
	}

	return false
}

// FilterJSONDataForUser returns only the vulnerabilities the user is allowed to view.
func FilterJSONDataForUser(jsonData []JSONData, email string, isGlobal bool) []JSONData {
	if isGlobal || len(jsonData) == 0 {
		return jsonData
	}

	filtered := make([]JSONData, 0, len(jsonData))
	for _, item := range jsonData {
		if CanViewVulnerability(item, email, isGlobal) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func GetVulnerabilityIds(isGlobal bool, email string, ids []uint) ([]uint, error) {
	var accessibleIds []uint

	query := db.Table("json_data").
		Select("json_data.id").
		Where("json_data.id IN ?", ids)

	if !isGlobal {
		query = query.Where(
			db.Where("json_data.project_id IN (?) AND deleted_at IS NULL",
				db.Table("project_data").
					Select("id").
					Where("client_email = ? OR hacker_name = ?", email, email),
			).Or("json_data.found_by = ? AND deleted_at IS NULL", email),
		)
	}

	err := query.Pluck("json_data.id", &accessibleIds).Error
	if err != nil {
		return nil, err
	}

	if !isGlobal && len(accessibleIds) > 0 {
		var records []JSONData
		if err := db.Where("id IN ?", accessibleIds).Find(&records).Error; err != nil {
			return nil, err
		}

		filtered := make([]uint, 0, len(records))
		for _, record := range records {
			if CanViewVulnerability(record, email, false) {
				filtered = append(filtered, record.ID)
			}
		}
		accessibleIds = filtered
	}

	return accessibleIds, nil
}

func SearchVulnerabilities(globalViewer bool, email string, isAdmin bool, params VulnerabilitySearchParams) ([]MinimalJSONData, error) {
	var jsonData []JSONData

	query := db.Preload("Project").
		Select(`
            json_data.*,
            json_extract(json_data.vulnerability, '$.title') as vulnerability_title,
            json_extract(json_data.vulnerability, '$.isPublicFacing') as vulnerability_isPublicFacing,
            json_extract(json_data.vulnerability, '$.criticality') as vulnerability_criticality,
            json_extract(json_data.vulnerability, '$.date') as vulnerability_date,
            json_extract(json_data.vulnerability, '$.visibility') as vulnerability_visibility,
            json_extract(json_data.vulnerability, '$.category') as vulnerability_category
        `).
		Where("json_data.deleted_at IS NULL").
		Order("json_data.created_at desc")

	if !isAdmin {
		query = query.Joins("INNER JOIN accessible_vulnerabilities ON accessible_vulnerabilities.id = json_data.id")
		if globalViewer {
			query = query.Where(
				db.Where(`accessible_vulnerabilities.visibility IN ('published', 'public')`).
					Or("accessible_vulnerabilities.assigned_to = ?", email).
					Or("accessible_vulnerabilities.found_by = ?", email).
					Or("',' || COALESCE(accessible_vulnerabilities.client_email, '') || ',' LIKE ?", "%,"+email+",%").
					Or("',' || COALESCE(accessible_vulnerabilities.hacker_name, '') || ',' LIKE ?", "%,"+email+",%"),
			)
		} else {
			query = query.Where(
				db.Where("accessible_vulnerabilities.assigned_to = ?", email).
					Or("accessible_vulnerabilities.found_by = ?", email).
					Or("',' || COALESCE(accessible_vulnerabilities.client_email, '') || ',' LIKE ?", "%,"+email+",%").
					Or("',' || COALESCE(accessible_vulnerabilities.hacker_name, '') || ',' LIKE ?", "%,"+email+",%"),
			)
		}
	}

	// Text search across title, category, evidence, remediation, endpoint, project name, foundBy, and ID
	if params.Query != "" {
		q := "%" + params.Query + "%"
		query = query.Where(
			db.Where("json_extract(json_data.vulnerability, '$.title') LIKE ?", q).
				Or("json_extract(json_data.vulnerability, '$.category') LIKE ?", q).
				Or("json_extract(json_data.vulnerability, '$.evidence') LIKE ?", q).
				Or("json_extract(json_data.vulnerability, '$.remediation') LIKE ?", q).
				Or("json_extract(json_data.vulnerability, '$.endpoint') LIKE ?", q).
				Or("json_data.found_by LIKE ?", q).
				Or("CAST(json_data.id AS TEXT) LIKE ?", q).
				Or("json_data.project_id IN (SELECT id FROM project_data WHERE project_name LIKE ?)", q),
		)
	}

	// Filter by criticality
	if len(params.Criticality) > 0 {
		query = query.Where("LOWER(json_extract(json_data.vulnerability, '$.criticality')) IN ?", params.Criticality)
	}

	// Filter by status
	if len(params.Status) > 0 {
		query = query.Where("json_data.status IN ?", params.Status)
	}

	// Filter by location
	if params.Location == "public" {
		query = query.Where("json_extract(json_data.vulnerability, '$.isPublicFacing') = true")
	} else if params.Location == "internal" {
		query = query.Where("json_extract(json_data.vulnerability, '$.isPublicFacing') = false")
	}

	// Hide closed (Resolved / Rejected)
	if params.HideClosed {
		query = query.Where("json_data.status NOT IN ?", []string{"Resolved", "Rejected"})
	}

	// Filter by year (based on vulnerability date field)
	if params.Year != "" {
		query = query.Where("json_extract(json_data.vulnerability, '$.date') LIKE ?", params.Year+"%")
	}

	// Assigned to me: match only the assignedTo field from the vulnerability JSON
	if params.AssignedToMe {
		if isAdmin {
			query = query.Where("LOWER(json_extract(json_data.vulnerability, '$.assignedTo')) = LOWER(?)", email)
		} else {
			query = query.Where("LOWER(accessible_vulnerabilities.assigned_to) = LOWER(?)", email)
		}
	}

	err := query.Find(&jsonData).Error
	if err != nil {
		return nil, err
	}

	filtered := FilterJSONDataForUser(jsonData, email, globalViewer || isAdmin)
	return minifiedVulnerabilityJSON(filtered), nil
}

func GetAdjacentVulnerabilityIDs(currentID uint, globalViewer bool, email string, isAdmin bool) (*uint, *uint, error) {
	var current struct {
		ID        uint
		CreatedAt time.Time
	}

	if err := db.Table("json_data").
		Select("id, created_at").
		Where("id = ? AND deleted_at IS NULL", currentID).
		Take(&current).Error; err != nil {
		return nil, nil, err
	}

	buildBaseQuery := func() *gorm.DB {
		query := db.Table("json_data").
			Select("json_data.id").
			Where("json_data.deleted_at IS NULL")

		if isAdmin {
			return query
		}

		query = query.Joins("INNER JOIN accessible_vulnerabilities ON accessible_vulnerabilities.id = json_data.id")
		if globalViewer {
			return query.Where(`
                accessible_vulnerabilities.visibility IN ('published', 'public') OR
                accessible_vulnerabilities.assigned_to = ? OR
                accessible_vulnerabilities.found_by = ? OR
                ',' || COALESCE(accessible_vulnerabilities.client_email, '') || ',' LIKE ? OR
                ',' || COALESCE(accessible_vulnerabilities.hacker_name, '') || ',' LIKE ?
            `, email, email, "%,"+email+",%", "%,"+email+",%")
		}

		return query.Where(`
                accessible_vulnerabilities.assigned_to = ? OR
                accessible_vulnerabilities.found_by = ? OR
                ',' || COALESCE(accessible_vulnerabilities.client_email, '') || ',' LIKE ? OR
                ',' || COALESCE(accessible_vulnerabilities.hacker_name, '') || ',' LIKE ?
            `, email, email, "%,"+email+",%", "%,"+email+",%")
	}

	var previousID *uint
	var nextID *uint

	var previous struct {
		ID uint
	}
	prevErr := buildBaseQuery().
		Where("(json_data.created_at < ?) OR (json_data.created_at = ? AND json_data.id < ?)", current.CreatedAt, current.CreatedAt, current.ID).
		Order("json_data.created_at DESC, json_data.id DESC").
		Limit(1).
		Take(&previous).Error
	if prevErr == nil {
		previousID = &previous.ID
	} else if !errors.Is(prevErr, gorm.ErrRecordNotFound) {
		return nil, nil, prevErr
	}

	var next struct {
		ID uint
	}
	nextErr := buildBaseQuery().
		Where("(json_data.created_at > ?) OR (json_data.created_at = ? AND json_data.id > ?)", current.CreatedAt, current.CreatedAt, current.ID).
		Order("json_data.created_at ASC, json_data.id ASC").
		Limit(1).
		Take(&next).Error
	if nextErr == nil {
		nextID = &next.ID
	} else if !errors.Is(nextErr, gorm.ErrRecordNotFound) {
		return nil, nil, nextErr
	}

	return previousID, nextID, nil
}

func minifiedVulnerabilityJSON(jsonData []JSONData) []MinimalJSONData {
	var minifiedList []MinimalJSONData
	for _, item := range jsonData {
		var minifiedVuln MinimalJSONData
		jsonVuln, _ := json.Marshal(item)
		json.Unmarshal(jsonVuln, &minifiedVuln)
		minifiedList = append(minifiedList, minifiedVuln)
	}
	return minifiedList
}

type VulnerabilityData struct {
	Category    string `json:"category"`
	Criticality string `json:"criticality"`
}

// getJSONData retrieves JSON data from the database
func GetJSONData(id uint) (JSONData, error) {
	var jsonData JSONData
	result := db.First(&jsonData, id)
	return jsonData, result.Error
}

func GetProject(id uint) (ProjectData, error) {
	var project ProjectData
	result := db.First(&project, id)
	return project, result.Error
}

func GetProjectsFor(email string) ([]ProjectData, error) {
	var projects []ProjectData

	// Prepare the database query to find projects where the email is in the ClientEmail list
	// Use exact match with comma separators to avoid substring attacks
	db := db.Where("(',' || client_email || ',') LIKE ?", "%,"+email+",%").
		Or("(',' || hacker_name || ',') LIKE ?", "%,"+email+",%")

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

func FindNonAvailablePersons(excludeProjectID uint, dateFrom, dateTo string) ([]string, error) {
	var assessmentJSONs []AssessmentJSON
	var emailsMap = make(map[string]bool) // Use a map to store unique emails
	var uniqueEmails []string

	// Build the initial query
	query := db.Model(&AssessmentJSON{}).Where("id != ?", excludeProjectID)

	// Apply date filters using the JSON_EXTRACT function in SQLite
	if dateFrom != "" {
		query = query.Where("date(json_extract(assessment, '$.dateFrom')) >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("date(json_extract(assessment, '$.dateTo')) <= ?", dateTo)
	}

	// Execute the query
	if err := query.Find(&assessmentJSONs).Error; err != nil {
		return nil, err
	}

	// Loop through the results and extract hacker emails
	for _, aJSON := range assessmentJSONs {
		var assessment models.Assessment
		if err := json.Unmarshal(aJSON.Assessment, &assessment); err != nil {
			return nil, err // Handle JSON unmarshal error
		}

		for _, hacker := range assessment.Hackers {
			emailsMap[hacker.Email] = true
		}
	}

	// Convert the map keys to a slice
	for email := range emailsMap {
		uniqueEmails = append(uniqueEmails, email)
	}

	return uniqueEmails, nil
}

func GetProjectVulnerabilities(projectID uint, dateFrom, dateTo string) ([]JSONData, error) {
	var jsonData []JSONData

	query := db.Where("project_id = ?", projectID)

	if dateFrom != "" {
		query = query.Where("date(json_extract(vulnerability, '$.date')) >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("date(json_extract(vulnerability, '$.date')) <= ?", dateTo)
	}

	err := query.Order("created_at desc").Find(&jsonData).Error

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
