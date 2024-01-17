package routes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"

	"prism/database"
	"strconv"
)

func GetProjects(c *gin.Context) {
	// Get the query parameter
	query := c.Query("query")

	projects, err := database.GetProjects(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func HandleProjectPut(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	var updatedProjectData database.ProjectData
	// Step 2: Bind incoming JSON data to the struct.
	if err := c.ShouldBindJSON(&updatedProjectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check that projectID matches updatedProjectData.ID
	if uint(projectID) != updatedProjectData.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID mismatch"})
		return
	}

	err = database.UpdateProject(&updatedProjectData)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully"})
}

func HandleProjectPost(c *gin.Context) {
	var projectData database.ProjectData
	if err := c.ShouldBindJSON(&projectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.CreateProject(&projectData)
	c.JSON(http.StatusCreated, gin.H{"id": projectData.ID})
}

func GetProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	id := uint(projectID)

	project, _ := database.GetProject(id)
	c.JSON(http.StatusOK, project)
}

func GetProjectVulnerabilitiesForProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	vulnerabilites, err := database.GetProjectVulnerabilities(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vulnerabilites)
}

func GetProjectVulnerabilitiesTotal(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	total, err := database.CountProjectVulnerabilities(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_vulnerabilities": total})
}

func DeleteProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	err = database.DeleteProjectAndAssets(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
