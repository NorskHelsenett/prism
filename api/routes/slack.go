package routes

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"

	"prism/config"
)

type ChannelData struct {
	Channels   []slack.Channel
	LastUpdate time.Time
}

var (
	instance *ChannelData
	once     sync.Once
	mux      sync.Mutex
)

func GetChannelDataInstance(api *slack.Client) *ChannelData {
	mux.Lock()
	defer mux.Unlock()

	once.Do(func() {
		instance = &ChannelData{}
		updateChannels(api)
	})

	if time.Since(instance.LastUpdate) > 60*time.Second {
		updateChannels(api)
	}

	return instance
}

func updateChannels(api *slack.Client) {
	var allChannels []slack.Channel
	cursor := ""

	for {
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			Limit:           100,
			Types:           []string{"public_channel"},
			ExcludeArchived: true,
		}

		channels, nextCursor, err := api.GetConversations(params)
		if err != nil {
			break // Handle error appropriately in production code
		}
		allChannels = append(allChannels, channels...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	instance.Channels = allChannels
	instance.LastUpdate = time.Now()
}

func GetSlackChannels(c *gin.Context) {

	slackAPI := slack.New(config.AppConfig.Slack.Token)

	query := c.Query("query")
	channelData := GetChannelDataInstance(slackAPI)
	result := channelData.SearchChannels(query)
	c.JSON(200, result)

}

func (cd *ChannelData) SearchChannels(query string) []slack.Channel {
	var result []slack.Channel
	for _, channel := range cd.Channels {
		if strings.Contains(channel.Name, query) {
			result = append(result, channel)
		}
	}
	return result
}
