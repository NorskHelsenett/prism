package event

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
	"strconv"
	"encoding/json"

	"prism/config"
	"prism/database"
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


func handleEvent(event database.EventQueue) {
		appConfig, _ := config.LoadConfig()
    finding,_ := database.GetJSONData(event.TableID)

		var vulnData Vulnerability
    err := json.Unmarshal(finding.Vulnerability, &vulnData)
    if err != nil {
        log.Printf("Error unmarshaling JSON data: %v", err)
        return
    }

		var data VulnerabilityData
		data.Vulnerability = vulnData
		data.URL = appConfig.Cors.Origin +"/vulnerability/" + strconv.FormatUint(uint64(finding.ID), 10) + "/view"
		var imageUrl string

		userID, _ := findUserIDByEmail(finding.FoundBy)

		if userID == "" {
			data.FoundBy = finding.FoundBy
		} else {
			data.FoundBy = userID
		}

		switch data.Vulnerability.Criticality {
		case "information":
				imageUrl = "https://imgur.com/kmNFZ6m.png" // Replace with actual URL
		case "low":
				imageUrl = "https://imgur.com/4LnvMDk.png" // Replace with actual URL
		case "medium":
				imageUrl = "https://imgur.com/4ztQrRz.png" // Replace with actual URL
		case "high":
				imageUrl = "https://imgur.com/2YCkzIc.png" // Replace with actual URL
		case "critical":
				imageUrl = "https://i.imgur.com/IlXt8Ds.png" // Replace with actual URL
		default:
				imageUrl = "https://imgur.com/kmNFZ6m.png" // Default image URL
		}

		data.ImageUrl = imageUrl

    if err := sendSlackMessage(data); err != nil {
        log.Printf("Failed to send Slack message: %v", err)
    } else {
			// Mark event as processed
			database.SetEventProcessed(&event)
		}
}

func PollEventQueue() {
	appConfig, _ := config.LoadConfig()
	log.Println("Starting polling for queue events")
	for {
		eventsPtr,err := database.GetOpenEvents()
		if err != nil {
			log.Println(err)
		}

		events := *eventsPtr

		for _, event := range events {
				go handleEvent(event)
		}

		time.Sleep(time.Duration(appConfig.Events.Interval) * time.Second)
	}
}

