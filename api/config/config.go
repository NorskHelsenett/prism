package config

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

// OIDCConfig represents the OIDC specific configuration.
type OIDCConfig struct {
    ClientID    string `yaml:"clientID"`
    ClientSecret string `yaml:"clientSecret"`
    RedirectURI string `yaml:"redirectUri"`
    ProviderURI string `yaml:"providerUri"`
}

type Database struct {
    Path string `yaml:"path"`
}

type Cors struct {
    Origin string `yaml:"origin"`
}

// Config represents the application configuration.
type Config struct {
    OIDC OIDCConfig `yaml:"oidc"`
    Cors Cors `yaml:"cors"`
    Database Database `yaml:"database"`
    Admins []string `yaml:"admins"`
    Events Events `yaml:"events"`
    Slack Slack `yaml:"slack"`
}

type Slack struct {
    Enable       bool `yaml:"enable"`
    Token        string `yaml:"token"`
    WebhookUrl   string `yaml:"webhookUrl"`
}

type Events struct {
    Interval int `yaml: "interval"`
}

// LoadConfig reads configuration from a YAML file specified by the CONFIG_PATH environment variable.
// When the env variable is empty, it will look for it in current directory path.
func LoadConfig() (*Config, error) {
    configPath := os.Getenv("CONFIG_PATH")

    if configPath == "" {
        cwd, err := os.Getwd()
        if err != nil {
            return nil, fmt.Errorf("error getting current working directory: %w", err)
        }
        configPath = filepath.Join(cwd, "config.yaml")
    }

    if configPath == "" {
        return nil, fmt.Errorf("CONFIG_PATH environment variable is required")
    }

    configFile, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("error reading config file: %w", err)
    }

    var config Config
    err = yaml.Unmarshal(configFile, &config)
    if err != nil {
        return nil, fmt.Errorf("error parsing config file: %w", err)
    }

    // Set default value for Interval if it's zero
    if config.Events.Interval == 0 {
        config.Events.Interval = 60
    }

    return &config, nil
}

