package routes

import (
    "github.com/gin-gonic/gin"
		"github.com/google/uuid"
    "net/http"
		"os"
		"path/filepath"

		"prism/config"
)

func HandleBlobUpload(c *gin.Context) {
    // Multipart form
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    files := form.File["image"] // "image" is the field name in your FormData
    var filenames []string
    for _, file := range files {
        // Generate a unique filename, for example using a UUID:
        filename := generateUniqueFilename(file.Filename)
				appConfig, _ := config.LoadConfig()

        filepath := filepath.Join(appConfig.Database.Path + "/images", filename)

        // Save the file
        if err := c.SaveUploadedFile(file, filepath); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
				filenames = append(filenames, filename)
    }

    c.JSON(http.StatusOK, gin.H{"fileNames": filenames})
}

func HandleBlobDelete(c *gin.Context) {
    filename := c.Param("filename") // Get the filename from the URL parameter

    // Construct the full path to the image file
    appConfig, _ := config.LoadConfig()
    filepath := filepath.Join(appConfig.Database.Path + "/images", filename)

    // Check if the file exists
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }

    // Attempt to delete the file
    if err := os.Remove(filepath); err != nil {
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
