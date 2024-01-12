package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"prism/config"
)

type Vulnerability struct {
	Title        string `json:"title"`
	Criticality  string `json:"criticality"`
	Category     string `json:"category"`
}

type VulnerabilityData struct {
	Vulnerability Vulnerability `json:"Vulnerability"`
	FoundBy       string
	ImageUrl      string
	URL           string
}

func sendSlackMessage(data VulnerabilityData) error {
		appConfig, _ := config.LoadConfig()
		if appConfig.Slack.Enable == false {
			return fmt.Errorf("Slack is disabled")
		}

		templateFilePath := os.Getenv("SLACK_PATH")

    templateString, err := readTemplateFile(templateFilePath)
    if err != nil {
        return err
    }

    tmpl, err := template.New("slackMessage").Parse(templateString)
    if err != nil {
        return err
    }

    var msgBuffer bytes.Buffer
    if err := tmpl.Execute(&msgBuffer, data); err != nil {
        return err
    }

    // Set up the HTTP request
    url := appConfig.Slack.WebhookUrl // Replace with your webhook URL if using a webhook
    response, err := http.Post(url, "application/json", &msgBuffer)
    if err != nil {
        log.Printf("Error making API call: %v", err)
        return err
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        log.Printf("Non-OK HTTP status: %d", response.StatusCode)
    }

    return nil
}

func readTemplateFile(filePath string) (string, error) {
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    return string(content), nil
}

func findUserIDByEmail(email string) (string, error) {
		appConfig, _ := config.LoadConfig()

    slackURL := fmt.Sprintf("https://slack.com/api/users.lookupByEmail?email=%s", url.QueryEscape(email))

    req, err := http.NewRequest("GET", slackURL, nil)
    if err != nil {
        return "", err
    }

		token := appConfig.Slack.Token
    req.Header.Add("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Ok    bool `json:"ok"`
        User struct {
            ID string `json:"id"`
        } `json:"user"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    if !result.Ok {
        return email, fmt.Errorf("failed to find user with email %s", email)
    }

    return result.User.ID, nil
}