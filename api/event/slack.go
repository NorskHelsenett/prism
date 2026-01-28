package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/slack-go/slack"
	"prism/config"
)

type Vulnerability struct {
	Title       string `json:"title"`
	Criticality string `json:"criticality"`
	Category    string `json:"category"`
}

type VulnerabilityData struct {
	Vulnerability Vulnerability `json:"Vulnerability"`
	FoundBy       string
	ImageUrl      string
	URL           string
}

func (v *Vulnerability) truncateTitle(maxLength int) {
	if len(v.Title) > maxLength {
		v.Title = v.Title[:maxLength]
	}
}

func (v *Vulnerability) replaceQuotes() {
	v.Title = strings.ReplaceAll(v.Title, "\"", " ")
}

func sendSlackMessage(data VulnerabilityData, channel string) (string, error) {

	// Check if the channel is empty
	if channel == "" {
		return "", fmt.Errorf("Channel is set to empty")
	}

	templateFilePath := os.Getenv("SLACK_PATH")
	templateString, err := readTemplateFile(templateFilePath)
	if err != nil {
		return "", err
	}

	data.Vulnerability.replaceQuotes()
	data.Vulnerability.truncateTitle(50)

	tmpl, err := template.New("slackMessage").Parse(string(templateString))
	if err != nil {
		return "", err
	}

	var msgBuffer bytes.Buffer
	if err := tmpl.Execute(&msgBuffer, data); err != nil {
		return "", err
	}

	// Unmarshal the blocks using the custom unmarshaler
	blocks, err := unmarshalBlocks(msgBuffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("Error unmarshaling blocks: %v", err)
	}

	api := slack.New(config.AppConfig.Slack.Token)

	channelID, timestamp, err := api.PostMessage(
		channel,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		return "", err
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)

	return timestamp, nil
}

func unmarshalBlocks(data []byte) ([]slack.Block, error) {
	var rawBlocks []json.RawMessage
	wrapper := struct {
		Blocks *[]json.RawMessage `json:"blocks"`
	}{Blocks: &rawBlocks}

	err := json.Unmarshal(data, &wrapper)
	if err != nil {
		return nil, err
	}

	var blocks []slack.Block
	for _, rawBlock := range rawBlocks {
		var blockType struct {
			Type string `json:"type"`
		}
		err := json.Unmarshal(rawBlock, &blockType)
		if err != nil {
			return nil, err
		}

		var block slack.Block
		switch blockType.Type {
		case "section":
			block = new(slack.SectionBlock)
		// Add other block types here
		default:
			continue
		}

		err = json.Unmarshal(rawBlock, block)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

func readTemplateFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func findUserIDByEmail(email string) (string, error) {
	api := slack.New(config.AppConfig.Slack.Token)

	user, err := api.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to find user with email %s: %v", email, err)
	}

	return user.ID, nil
}
