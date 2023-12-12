package database

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/datatypes"
    "errors"
    "encoding/json"

    "prism/config"
)

// JSONData is a simple model for storing JSON data
type JSONData struct {
    gorm.Model
    Vulnerability datatypes.JSON
    FoundBy string
}

type UserData struct {
    gorm.Model
    Email string
    Name string
    Picture string
}

type Vulnerability struct {
    Criticality string `json:"criticality"`
    Category string `json:"category"`
}

var db *gorm.DB

func InitDB() {
    appConfig, _ := config.LoadConfig()

    var err error
    db, err = gorm.Open(sqlite.Open(appConfig.Database.Path +"/prism.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect to the database")
    }

    // Migrate the schema
    db.AutoMigrate(&JSONData{})
    db.AutoMigrate(&UserData{})
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

func AllVulnerabilities() ([]JSONData, error){
    var jsonData []JSONData
    result := db.Find(&jsonData)
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
