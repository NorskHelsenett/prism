package routes

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prism/database"
)

func GetBlob(c *gin.Context) {
	filename := c.Param("filename")

	imgData, err := database.GetImage(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving file"})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", imgData.Data)
}

func HandleBlobUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["image"]
	var filenames []string
	for _, file := range files {
		filename := generateUniqueFilename(file.Filename)

		// Read file data
		fileData, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer fileData.Close()

		data, err := io.ReadAll(fileData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Save the file data to the database
		if err := database.SaveImage(filename, data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filenames = append(filenames, filename)
	}

	c.JSON(http.StatusOK, gin.H{"fileNames": filenames})
}

func HandleBlobDelete(c *gin.Context) {
	filename := c.Param("filename") // Get the filename from the URL parameter

	// Attempt to delete the image record from the database
	err := database.DeleteImage(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func generateUniqueFilename(originalName string) string {
	newUUID := uuid.New()
	extension := filepath.Ext(originalName)
	return newUUID.String() + extension
}
