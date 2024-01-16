package routes

import (
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"github.com/gin-gonic/gin"

	"net/http"
	"path/filepath"
	"io"
	"os"
	"log"
	"bytes"

	"prism/config"
	"prism/database"
)

func ImportData(c *gin.Context) {
	// Receive the file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// File size limitation (100 MB)
	if file.Size > 100*1024*1024 {
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
			"error": "Failed to open uploaded SQLite database directly with GORM",
			"detail": err.Error(),
			"file": file.Filename,
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

	// Load app config for database path
	appConfig, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configuration"})
		return
	}
	dbFilePath := filepath.Join(appConfig.Database.Path, "prism.db")

	// Save the file directly
	err = c.SaveUploadedFile(file, dbFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded, verified, and saved successfully"})
}

// ExportAllData handles the downloading of the prism.db database file.
func ExportAllData(c *gin.Context) {
	appConfig, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configuration"})
		return
	}

	dbFilePath := filepath.Join(appConfig.Database.Path, "prism.db")
	c.Header("Content-Disposition", "attachment; filename=prism.db")
	c.FileAttachment(dbFilePath, "prism.db")
}

func GetSettings(c *gin.Context) {
	settings, err := database.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func PostSettings(c *gin.Context){
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