package routes

import (
	"encoding/json"
	"net/http"
	"prism/database"
	"prism/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Default values for pagination
const (
	DefaultPage     = 1
	DefaultPageSize = 100
)

func NewAssassment(c *gin.Context) {
	email, _ := c.Get("email")
	var assessment models.Assessment
	assessment.Responsible = email.(string)

	if err := c.ShouldBindJSON(&assessment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := assessment.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := database.PersistAssessment(assessment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	assessment.ID = id

	c.JSON(http.StatusOK, assessment)
}

func PatchAssessmentsHandler(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
	}

	// Convert string ID to uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
	}

	// Retrieve the existing assessment
	assessmentJSON, err := database.RetrieveAssessment(uint(id))
	if err != nil {
			if strings.Contains(err.Error(), "not found") {
					c.AbortWithStatus(http.StatusNotFound)
					return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting assessment"})
			return
	}

	var assessment models.Assessment
	if err := json.Unmarshal(assessmentJSON.Assessment, &assessment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse assessment data"})
			return
	}

	// Create a map to hold the patch data
	var patchData map[string]interface{}
	if err := c.ShouldBindJSON(&patchData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
	}

	// Apply patch data to the assessment
	if color, ok := patchData["color"].(string); ok {
			assessment.Color = color
	}

	if title, ok := patchData["title"].(string); ok { assessment.Title = title }

	if startDate, ok := patchData["dateFrom"].(string); ok {
		assessment.DateFrom = startDate
	}
	if endDate, ok := patchData["dateTo"].(string); ok {
		assessment.DateTo = endDate
	}

	if hackers, ok := patchData["hackers"].([]interface{}); ok {
		// Use a map to track unique emails
		uniqueEmails := make(map[string]bool)
		var uniqueHackers []models.Hacker
		
		for _, hacker := range hackers {
			if hackerMap, ok := hacker.(map[string]interface{}); ok {
				if email, ok := hackerMap["email"].(string); ok {
					// Only add if this email hasn't been seen before
					if !uniqueEmails[email] {
						uniqueEmails[email] = true
						uniqueHackers = append(uniqueHackers, models.Hacker{Email: email})
					}
				}
			}
		}
		
		assessment.Hackers = uniqueHackers
	}

	// Handle projects update - only store IDs
	if projectsData, ok := patchData["projects"].([]interface{}); ok {
		var projects []models.Project
		for _, projectData := range projectsData {
			if projectMap, ok := projectData.(map[string]interface{}); ok {
				if id, ok := projectMap["id"].(float64); ok { // JSON numbers come as float64
					projects = append(projects, models.Project{Id: uint(id)})
				}
			}
		}
		assessment.Projects = projects
	}

	// Validate the modified assessment
	if err := assessment.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
	}

	// Update the assessment in the database
	err = database.UpdateAssassment(assessment, uint(id))
	if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update assessment"})
			return
	}

	assessment.ID = assessmentJSON.ID
	c.JSON(http.StatusOK, assessment)
}

func DeleteAssessmentsHandler(c *gin.Context) {
	idStr := c.Param("id") // Get ID as string
	if idStr != "" {
		// Convert string ID to uint
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		err = database.DeleteAssessment(uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete assessment"})
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted successfully"})
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "id is required"})
	}
}

func PutAssessmentsHandler(c *gin.Context) {
	idStr := c.Param("id") // Get ID as string
	if idStr != "" {
		// Convert string ID to uint
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		var assessment models.Assessment

		if err := c.ShouldBindJSON(&assessment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := assessment.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		error := database.UpdateAssassment(assessment, uint(id))
		if error != nil {
			c.AbortWithStatus(500)
			return
		}

		c.JSON(http.StatusOK, assessment)
	}
}

func RetrieveAssessmentsHandler(c *gin.Context) {

	idStr := c.Param("id") // Get ID as string
	if idStr != "" {
		// Convert string ID to uint
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		assessmentJSON, err := database.RetrieveAssessment(uint(id))
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.AbortWithStatus(http.StatusNotFound)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting assessment"})
			return
		}

		var assessment models.Assessment
		if err := json.Unmarshal(assessmentJSON.Assessment, &assessment); err != nil {
			// Handle JSON unmarshal error, maybe log it or return an HTTP error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse assessment data"})
			return
		}

		// Iterate over assessment.Projects by index
		for j := range assessment.Projects {
			// Use the index to directly reference and modify each project
			name, err := database.PopulateProjectName(assessment.Projects[j].Id)
			if err != nil {
				// Handle the error, perhaps log it or return an HTTP error response
				// For simplicity, let's continue to the next project
				continue
			}
			assessment.Projects[j].Name = name
		}
		assessment.ID = assessmentJSON.ID
		c.JSON(http.StatusOK, assessment)
		return
	}

	// Get the current time
	now := time.Now()

	// Set the default startDate to the first day of the current month
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Set the default endDate to 3 months into the future
	endDate := startDate.AddDate(0, 12, 0)

	// Read the startDate and endDate from query parameters, use default values if not provided
	startDateStr := c.DefaultQuery("startDate", startDate.Format("2006-01-02"))
	endDateStr := c.DefaultQuery("endDate", endDate.Format("2006-01-02"))
	pageStr := c.DefaultQuery("page", string(DefaultPage))
	pageSizeStr := c.DefaultQuery("pageSize", string(DefaultPageSize))

	// Parse dates
	// startDate, err := time.Parse("2006-01-02", startDateStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate format"})
	// 	return
	// }

	// endDate, err := time.Parse("2006-01-02", endDateStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate format"})
	// 	return
	// }

	// Parse pagination parameters
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// Ensure page and pageSize are within reasonable bounds or set to defaults
	if page < 1 {
		page = DefaultPage
	}
	if pageSize <= 0 || pageSize > 100 { // Assuming a max of 100 items per page for sanity
		pageSize = DefaultPageSize
	}

	// Fetch the assessments
	assessments, err := database.RetrieveAssessments(startDateStr, endDateStr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve assessments"})
		return
	}

	var modelAssessments []models.Assessment
	for _, assessmentJSON := range assessments {
		var assessment models.Assessment
		if err := json.Unmarshal(assessmentJSON.Assessment, &assessment); err != nil {
			// Handle JSON unmarshal error, maybe log it or return an HTTP error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse assessment data"})
			return
		}

		// Iterate over assessment.Projects by index
		for j := range assessment.Projects {
			// Use the index to directly reference and modify each project
			name, err := database.PopulateProjectName(assessment.Projects[j].Id)
			if err != nil {
				// Handle the error, perhaps log it or return an HTTP error response
				// For simplicity, let's continue to the next project
				continue
			}
			assessment.Projects[j].Name = name
		}

		if assessment.Hackers == nil {
			assessment.Hackers = []models.Hacker{}
	}

		assessment.ID = assessmentJSON.ID
		modelAssessments = append(modelAssessments, assessment)
	}

	//marshall assessments.Assessment to models.Assessments

	// Respond with the fetched data
	c.JSON(http.StatusOK, modelAssessments)
}

func FindNonAvailablePersons(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	// Retrieve optional dateFrom and dateTo query parameters
	dateFrom := c.Query("from")
	dateTo := c.Query("to")

	availablePersons, err := database.FindNonAvailablePersons(uint(id), dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, availablePersons)
}
