package routes

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"prism/config"
)

func ImportData(c *gin.Context) {
	// Load the application configuration
	appConfig, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configuration"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Specify the path where the database file should be saved.
	dbFilePath := filepath.Join(appConfig.Database.Path, "prism.db")

	// Save the file. You may want to add extra checks or a backup mechanism here.
	err = c.SaveUploadedFile(file, dbFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
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
