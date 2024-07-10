package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"prism/config"
	"prism/database"
	"prism/models"
	"prism/routes"
)

func EventQueues(c *gin.Context) {
	// Retrieve the limit from query parameter, default to 0 if not provided
	limitQuery := c.DefaultQuery("limit", "0")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		// Handle error if the limit is not an integer
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	events, err := database.GetAllEvents(limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func DeleteEventQueue(c *gin.Context) {
	// Extract the vulnerability ID from the URL parameters
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Call UpdateEvent with the ID
	err = database.DeleteEvent(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
}

func UpdateEventQueues(c *gin.Context) {
	// Extract the vulnerability ID from the URL parameters
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	statusStr := c.Param("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	status, err := strconv.ParseBool(statusStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status format"})
		return
	}

	// Call UpdateEvent with the ID
	err = database.UpdateEvent(uint(id), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func handleEvent(event database.EventQueue) error {
	settings, _ := database.GetSettings(false)

	if !settings.Slack.Enabled {
		return fmt.Errorf("slack is disabled")
	}

	finding, err := database.GetJSONData(event.TableID)

	if err != nil {
		return fmt.Errorf("error vulnerability not found: %v", err)
	}

	var vulnData Vulnerability
	err = json.Unmarshal(finding.Vulnerability, &vulnData)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	var data VulnerabilityData
	data.Vulnerability = vulnData
	data.URL = config.AppConfig.Cors.Origin + "/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view"
	var imageUrl string

	userID, _ := findUserIDByEmail(finding.FoundBy)

	if userID == "" {
		data.FoundBy = finding.FoundBy
	} else {
		data.FoundBy = userID
	}

	switch data.Vulnerability.Criticality {
	case "information":
		imageUrl = "https://imgur.com/kmNFZ6m.png"
	case "low":
		imageUrl = "https://imgur.com/4LnvMDk.png"
	case "medium":
		imageUrl = "https://imgur.com/4ztQrRz.png"
	case "high":
		imageUrl = "https://imgur.com/2YCkzIc.png"
	case "critical":
		imageUrl = "https://i.imgur.com/IlXt8Ds.png"
	default:
		imageUrl = "https://imgur.com/kmNFZ6m.png"
	}

	data.ImageUrl = imageUrl

	if finding.ProjectID != nil {
		log.Printf("ProjectID is empty")
	}

	slackChannel := settings.Slack.ChannelID

	if finding.ProjectID != nil {
		project, _ := database.GetProject(*finding.ProjectID)

		// Check if the channel is empty
		if project.SlackChannel == "" {
			project.SlackChannel = slackChannel
			if project.SlackChannel == "" {
				return fmt.Errorf("Channel is set to empty. Update it in settings.")
			}
		}
	}

	timestamp, err := sendSlackMessage(data, slackChannel)
	if err != nil { //timestamp comes from here
		return fmt.Errorf("Failed to send Slack message for channel %s: %v", slackChannel, err)
	} else {
		// Mark event as processed
		url := getUrlFor(slackChannel, timestamp, settings.Slack.Workspace)
		err = database.SetVulnerabilitySlackUrl(finding.ID, url)
		if err != nil {
			return fmt.Errorf("Failed to get URL for channel %w", err)
		}
	}
	return nil
}

func getUrlFor(channel string, timestamp string, workspace string) string {
	// Construct the URL
	// Replace the period in the timestamp with an empty string
	formattedTimestamp := strings.Replace(timestamp, ".", "", 1)
	return fmt.Sprintf("slack://%s.slack.com/archives/%s/p%s", workspace, channel, formattedTimestamp)
}

func sendBrowserNotification(event database.EventQueue) {

	if event.Kind != models.NewVulnerability {
		return
	}

	finding, err := database.GetJSONData(event.TableID)

	if err != nil {
		updateEvent(event, err)
		log.Printf("Error getting the event from table %s", err)
		return
	}

	var vulnData Vulnerability
	err = json.Unmarshal(finding.Vulnerability, &vulnData)
	if err != nil {
		updateEvent(event, err)
		log.Printf("Error unmarshelling the event %s", err)

		return
	}

	var data VulnerabilityData
	data.Vulnerability = vulnData
	data.URL = "/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view"

	usersToNotify := make(map[string]bool)
	var projectName string

	if finding.ProjectID != nil {
		project, _ := database.GetProject(*finding.ProjectID)
		projectName = project.ProjectName

		// Split the HackerName string by comma and add each name to the map
		for _, hacker := range strings.Split(project.HackerName, ",") {
			if hacker != "" { // Make sure the string is not empty
				usersToNotify[hacker] = true
			}
		}

		// Split the ClientEmail string by comma and add each email to the map
		for _, email := range strings.Split(project.ClientEmail, ",") {
			if email != "" { // Make sure the string is not empty
				usersToNotify[email] = true
			}
		}
	}

	// missing global users
	// missing manually subscribed to project notification users

	err = routes.SendMessage(projectName, "New vulnerability "+data.Vulnerability.Title, data.URL, finding.FoundBy, finding.FoundBy, usersToNotify)
	if err != nil {
		log.Printf("Error sending the message %s", err)
	}
	updateEvent(event, err)
}

func updateEvent(event database.EventQueue, err error) {
	if err != nil {
		// Log the error and mark the event as processed with an error
		log.Printf("Error processing event: %v", err)
		event.Error = fmt.Sprintf("Error processing event: %v", err)
		database.SetEventProcessed(&event)
	} else {
		// If no error, mark the event as processed normally
		database.SetEventProcessed(&event)
	}
}

func sendCommentsNotification(event database.EventQueue) {

	if event.Kind != models.NewComment {
		return
	}

	finding, err := database.GetJSONData(event.TableID)

	if err != nil {
		return
	}

	var vulnData Vulnerability
	err = json.Unmarshal(finding.Vulnerability, &vulnData)
	if err != nil {
		updateEvent(event, err)
		return
	}

	var data VulnerabilityData
	data.Vulnerability = vulnData

	vulnerability, _ := database.GetJSONData(finding.ID)

	var comments []models.Comment

	_ = json.Unmarshal([]byte(vulnerability.Comments), &comments)

	usersToNotify := make(map[string]bool)

	var lastComment models.Comment

	// Ensure we have at least one comment to initialize lastComment
	if len(comments) > 0 {
		lastComment = comments[0]
	}

	for _, comment := range comments {
		usersToNotify[comment.UserEmail] = true
		if lastComment.CreatedAt.Before(comment.CreatedAt) {
			lastComment = comment
		}
	}

	data.URL = "/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view#" + lastComment.ID

	// usersToNotify = append(usersToNotify, vulnerability.FoundBy)
	usersToNotify[vulnerability.FoundBy] = true

	if finding.ProjectID != nil {
		project, _ := database.GetProject(*finding.ProjectID)
		var projectUsers []string
		projectUsers = append(projectUsers, strings.Split(project.HackerName, ",")...)
		projectUsers = append(projectUsers, strings.Split(project.ClientEmail, ",")...)

		for _, user := range projectUsers {
			usersToNotify[user] = true
		}
	}

	var message string
	if len(lastComment.Text) > 50 {
		message = "💬 " + lastComment.Text[:50]
	} else {
		message = "💬 " + lastComment.Text
	}

	err = routes.SendMessage("PRISM", message, data.URL, finding.FoundBy, lastComment.UserEmail, usersToNotify)
	updateEvent(event, err)
}

func prepareAndSendSlackMessage(event database.EventQueue) {
	if event.Kind != models.NewVulnerability {
		return
	}

	err := handleEvent(event)
	if err != nil {
		// Log the error and mark the event as processed with an error
		log.Printf("Error processing event: %v", err)
		event.Error = fmt.Sprintf("Error processing event: %v", err)
		database.SetEventProcessed(&event)
	} else {
		// If no error, mark the event as processed normally
		database.SetEventProcessed(&event)
	}
}

func PollEventQueue() {
	log.Println("Starting polling for queue events")

	for {
		time.Sleep(time.Duration(config.AppConfig.Events.Interval) * time.Second)

		eventsPtr, err := database.GetOpenEvents()
		if err != nil {
			log.Println(err)
			return
		}

		events := *eventsPtr

		for _, event := range events {
			eventCopy := event
			go prepareAndSendSlackMessage(eventCopy)
			go sendBrowserNotification(eventCopy)
			go sendCommentsNotification(eventCopy)
		}
	}
}
