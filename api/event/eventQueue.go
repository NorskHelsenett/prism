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
	if err := json.Unmarshal(finding.Vulnerability, &vulnData); err != nil {
		updateEvent(event, err)
		log.Printf("Error unmarshalling the event %s", err)
		return
	}

	url := "/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view"

	recipients := []string{}
	var projectName string
	if finding.ProjectID != nil {
		project, _ := database.GetProject(*finding.ProjectID)
		projectName = project.ProjectName
		recipients = append(recipients, splitMembers(project.HackerName)...)
		recipients = append(recipients, splitMembers(project.ClientEmail)...)
	}

	err = routes.Dispatch(routes.DispatchRequest{
		Kind:            models.NotificationKindNewVuln,
		VulnerabilityID: finding.ID,
		ActorEmail:      finding.FoundBy,
		Recipients:      recipients,
		Title:           projectName,
		Body:            "New vulnerability " + vulnData.Title,
		URL:             url,
	})
	if err != nil {
		log.Printf("Error sending the message %s", err)
	}
	updateEvent(event, err)
}

// splitMembers parses one of the comma-separated email columns on
// ProjectData. The Dispatcher does its own trim+lowercase pass, but doing it
// here too keeps both halves of the code paths consistent and makes the
// recipient list easier to reason about in logs.
func splitMembers(commaSeparated string) []string {
	parts := strings.Split(commaSeparated, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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

// commentForEvent picks the comment that plausibly caused this event. The
// comments trigger fires on every UPDATE of the comments column — including
// deletions — so an event with no comments left, or whose newest comment
// clearly predates the event, was a deletion and must not notify anyone.
// Comment timestamps are server-set (routes.NewComment/UpdateComment), so the
// comparison is between two server clocks; the tolerance absorbs trigger
// timestamp granularity. A deletion within the tolerance window of the newest
// comment still re-notifies — accepted noise for a much rarer case.
func commentForEvent(event database.EventQueue, comments []models.Comment) (models.Comment, bool) {
	if len(comments) == 0 {
		return models.Comment{}, false
	}
	last := comments[0]
	for _, c := range comments {
		if last.CreatedAt.Before(c.CreatedAt) {
			last = c
		}
	}
	if event.CreatedAt.Sub(last.CreatedAt) > 2*time.Minute {
		return models.Comment{}, false
	}
	return last, true
}

func sendCommentsNotification(event database.EventQueue) {
	if event.Kind != models.NewComment {
		return
	}

	finding, err := database.GetJSONData(event.TableID)
	if err != nil {
		return
	}

	vulnerability, _ := database.GetJSONData(finding.ID)

	var comments []models.Comment
	_ = json.Unmarshal([]byte(vulnerability.Comments), &comments)

	lastComment, ok := commentForEvent(event, comments)
	if !ok {
		// The comments column changed but the newest comment predates the
		// event — a deletion, not a new comment. Don't re-notify.
		updateEvent(event, nil)
		return
	}

	recipients := []string{vulnerability.FoundBy}
	for _, comment := range comments {
		recipients = append(recipients, comment.UserEmail)
	}

	if finding.ProjectID != nil {
		project, _ := database.GetProject(*finding.ProjectID)
		recipients = append(recipients, splitMembers(project.HackerName)...)
		recipients = append(recipients, splitMembers(project.ClientEmail)...)
	}

	url := "/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view#" + lastComment.ID

	var message string
	if len(lastComment.Text) > 50 {
		message = "💬 " + lastComment.Text[:50]
	} else {
		message = "💬 " + lastComment.Text
	}

	err = routes.Dispatch(routes.DispatchRequest{
		Kind:            models.NotificationKindNewComment,
		VulnerabilityID: finding.ID,
		ActorEmail:      lastComment.UserEmail,
		Recipients:      recipients,
		Title:           "PRISM",
		Body:            message,
		URL:             url,
	})
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
