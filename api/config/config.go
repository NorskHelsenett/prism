package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// AppConfig holds the application configuration in-memory.
var AppConfig *Config

// OIDCConfig represents the OIDC specific configuration.
type OIDCConfig struct {
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	RedirectURI  string `yaml:"redirectUri"`
	ProviderURI  string `yaml:"providerUri"`
}

type Database struct {
	Path string `yaml:"path"`
}

type Cors struct {
	Origin string `yaml:"origin"`
}

// Config represents the application configuration.
type Config struct {
	OIDC     map[string]OIDCConfig `yaml:"oidc"`
	Cors     Cors                  `yaml:"cors"`
	Database Database              `yaml:"database"`
	Admins   []string              `yaml:"admins"`
	Events   Events                `yaml:"events"`
	Slack    Slack                 `yaml:"slack"`
	Roles    map[string]Role       `yaml:"-"`
}

type RolesFile struct {
	Roles map[string]Role `yaml:"roles"`
}

type Slack struct {
	Token      string `yaml:"token"`
	WebhookUrl string `yaml:"webhookUrl"`
}

type Events struct {
	Interval int `yaml: "interval"`
}

// Permission struct for parsing YAML
type Permission struct {
	Resource string   `yaml:"resource"`
	Action   []string `yaml:"action"`
}

// Role struct for parsing YAML
type Role struct {
	Description string       `yaml:"description"`
	Permissions []Permission `yaml:"permissions"`
}

func LoadRoles(rolesPath string) (map[string]Role, error) {
	rolesFile, err := os.ReadFile(rolesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading roles file: %w", err)
	}

	var roles RolesFile
	err = yaml.Unmarshal(rolesFile, &roles)
	if err != nil {
		return nil, fmt.Errorf("error parsing roles file: %w", err)
	}

	// Define all possible actions
	allActions := []string{"read", "write", "delete"}

	// Use a map to track unique resources
	uniqueResources := make(map[string]bool)

	// Iterate over roles and permissions to find unique resources
	for _, role := range roles.Roles {
		for _, perm := range role.Permissions {
			uniqueResources[perm.Resource] = true
		}
	}

	// Remove wildcard '*' and add specific resources
	for roleName, role := range roles.Roles {
		// Temporary map to collect permissions and check for deny attributes
		resourceActions := make(map[string][]string)

		// First pass to collect all permissions
		for _, perm := range role.Permissions {
			// Handle wildcards occurrences
			if perm.Resource == "*" {
				for resource := range uniqueResources {
					resourceActions[resource] = append(resourceActions[resource], perm.Action...)
				}
			} else {
				// Collect actions for specific resources
				resourceActions[perm.Resource] = append(resourceActions[perm.Resource], perm.Action...)
			}
		}

		// Filter to remove wildcards and to enforce deny attributes
		newPermissions := []Permission{}

		// Iterate to resolve deny overrides and filter out wildcard resources
		for resource, actions := range resourceActions {
			// Skip wildcard resources
			if resource == "*" {
				continue
			}

			var finalActions []string
			hasDeny := false

			// Check for deny attribute, which overrides any allow attributes
			for _, action := range actions {
				if action == "" {
					hasDeny = true
					break
				}
				if action != "*" && !contains(finalActions, action) {
					finalActions = append(finalActions, action)
				}
			}

			// If "*" is one of the actions and deny is not present, allow all actions
			if contains(actions, "*") && !hasDeny {
				finalActions = allActions
			}

			// Only add non-empty, non-denied permissions to the new list
			if !hasDeny {
				newPerm := Permission{
					Resource: resource,
					Action:   finalActions,
				}
				newPermissions = append(newPermissions, newPerm)
			}
		}

		// Update the role with resolved permissions excluding any wildcard resources
		updatedRole := Role{
			Description: role.Description,
			Permissions: newPermissions,
		}

		// Update the role in the main configuration
		roles.Roles[roleName] = updatedRole
	}

	return roles.Roles, nil
}

// LoadConfig reads configuration from a YAML file specified by the CONFIG_PATH environment variable.
// When the env variable is empty, it will look for it in current directory path.
func LoadConfig() error {
	configPath := os.Getenv("CONFIG_PATH")
	rolesPath := os.Getenv("ROLES_PATH")

	if rolesPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		rolesPath = filepath.Join(cwd, "roles.yaml")
	}

	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		configPath = filepath.Join(cwd, "config.yaml")
	}

	if configPath == "" {
		return fmt.Errorf("CONFIG_PATH environment variable is required")
	}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Set default value for Interval if it's zero
	if config.Events.Interval == 0 {
		config.Events.Interval = 60
	}

	// Now load separate roles
	roles, _ := LoadRoles(rolesPath)
	config.Roles = roles // Combine the roles into the main config

	AppConfig = &config // Set global config

	return nil
}

// Utility function to check if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
