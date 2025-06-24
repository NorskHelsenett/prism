package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"prism/config"
	"prism/database"
)

func CleanUpDatabase(c *gin.Context) {
	err := database.CleanUpDatabase()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error cleaning database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func ImportData(c *gin.Context) {
	// Receive the file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// File size limitation (300 MB)
	if file.Size > 300*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open the uploaded file"})
		return
	}
	defer src.Close()

	// Read the file header for SQLite format check
	header := make([]byte, 16)
	if _, err := io.ReadFull(src, header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the uploaded file header"})
		return
	}

	const sqliteHeader = "SQLite format 3\x00"
	if !bytes.Equal(header, []byte(sqliteHeader)) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Uploaded file is not a valid SQLite database header"})
		return
	}

	src.Seek(0, 0) // Reset the read pointer to the beginning

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "upload-*.db")
	if err != nil {
		log.Printf("Failed to create a temporary file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create a temporary file", "detail": err.Error()})
		return
	}
	defer os.Remove(tempFile.Name()) // Clean up
	defer tempFile.Close()

	// Copy to temp file and reset pointer
	if _, err = io.Copy(tempFile, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy to temporary file"})
		return
	}
	if _, err = tempFile.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file read pointer"})
		return
	}

	// Open temporary SQLite file with GORM to verify
	db, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":  "Failed to open uploaded SQLite database directly with GORM",
			"detail": err.Error(),
			"file":   file.Filename,
		})
		return
	}

	// Database integrity check with GORM
	var integrityCheck string
	err = db.Raw("PRAGMA integrity_check;").Scan(&integrityCheck).Error
	if err != nil || integrityCheck != "ok" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Database integrity check failed"})
		return
	}

	// Assuming db is your database connection object
	// and there is a function to get this object
	err = database.CloseConnection()
	if err != nil {
		// handle error
	}

	dbFilePath := filepath.Join(config.AppConfig.Database.Path, "prism.db")

	// Save the file directly
	err = c.SaveUploadedFile(file, dbFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	// Re-establish a new connection to the new database
	// You will need to implement the logic to reinitialize the DB connection
	// depending on how your application is structured
	database.InitDB()

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded, verified, and saved successfully"})
	// log.Fatalf("Exiting app due to new database written")
}

// ExportAllData handles the downloading of the prism.db database file.
func ExportAllData(c *gin.Context) {

	dbFilePath := filepath.Join(config.AppConfig.Database.Path, "prism.db")
	c.Header("Content-Disposition", "attachment; filename=prism.db")
	c.FileAttachment(dbFilePath, "prism.db")
}

func GetSettings(c *gin.Context) {
	settings, err := database.GetSettings(true)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func PostSettings(c *gin.Context) {
	var settings database.Settings

	// Parse the incoming JSON to newSettings
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	err := database.UpdateSettings(&settings)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// GetAllRoles retrieves all role names from the application configuration and
// returns them as an array of strings.
func GetAllRoles(c *gin.Context) {
	var roleNames []string

	// Assuming AppConfig is an instance of Config that has been initialized,
	// and it is globally accessible within this context.
	for roleName := range config.AppConfig.Roles {
		roleNames = append(roleNames, roleName)
	}

	c.JSON(http.StatusOK, roleNames)
}
