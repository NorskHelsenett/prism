package routes

import (
	"net/http"
	"strings"

	"prism/config"

	"github.com/gin-gonic/gin"
)

type OIDCProvider struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ClientConfig struct {
	APIEndpoint string         `json:"apiEndpoint"`
	Providers   []OIDCProvider `json:"providers"`
}

// HandleClientConfig returns the client configuration dynamically based on the server config
func HandleClientConfig(c *gin.Context) {
	// Build the list of configured OIDC providers
	var providers []OIDCProvider

	for providerType, providerConfig := range config.AppConfig.OIDC {
		// Use the name from the config, fallback to formatted type if not set
		name := providerConfig.Name
		if name == "" {
			name = formatProviderName(providerType)
		}
		providers = append(providers, OIDCProvider{
			Name: name,
			Type: providerType,
		})
	}

	// Determine the API endpoint from the CORS origin configuration
	apiEndpoint := config.AppConfig.Cors.Origin

	// If CORS origin is not set, try to infer from the request
	if apiEndpoint == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		apiEndpoint = scheme + "://" + c.Request.Host
	}

	clientConfig := ClientConfig{
		APIEndpoint: apiEndpoint,
		Providers:   providers,
	}

	c.JSON(http.StatusOK, clientConfig)
}

// formatProviderName converts provider type to a human-readable name
func formatProviderName(providerType string) string {
	// Convert to title case and handle special cases
	switch providerType {
	case "azure":
		return "Microsoft AD"
	case "gitlab":
		return "Helsegitlab"
	default:
		// Capitalize first letter
		if len(providerType) > 0 {
			return strings.ToUpper(providerType[:1]) + providerType[1:]
		}
		return providerType
	}
}
