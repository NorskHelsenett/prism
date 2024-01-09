package database

import (
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"os"
	"path/filepath"

	"prism/config"
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

// JSONData is a simple model for storing JSON data
type JSONData struct {
	gorm.Model
	Vulnerability datatypes.JSON
	FoundBy       string
	ProjectID     *uint        // Foreign key for ProjectData
	Project       *ProjectData // The associated project
}

type UserData struct {
	gorm.Model
	Email   string
	Name    string
	Picture string
}

type Vulnerability struct {
	Criticality string `json:"criticality"`
	Category    string `json:"category"`
	ProjectID   uint   `json:"projectID"`
}

var db *gorm.DB

func InitDB() {
	appConfig, _ := config.LoadConfig()

	var err error
	db, err = gorm.Open(sqlite.Open(appConfig.Database.Path+"/prism.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Migrate the schema
	db.AutoMigrate(&JSONData{})
	db.AutoMigrate(&UserData{})
	db.AutoMigrate(&ProjectData{})
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

func GetUserDataByEmail(email string) (*UserData, error) {
	var userData UserData
	result := db.Where("email = ?", email).First(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

// createJSONData saves new JSON data to the database
func CreateJSONData(jsonData *JSONData) {
	db.Create(jsonData)
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

	// Perform a search if a query is provided
	if query != "" {
		result := db.Where("project_name LIKE ?", "%"+query+"%").Find(&projects)
		return projects, result.Error
	}

	// Return all projects if no query is provided
	result := db.Find(&projects)

	return projects, result.Error
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
                    imagePath := filepath.Join("./tmp/images/", imageName)
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
