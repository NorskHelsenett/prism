package auth

import (
	"crypto/rand"
	"encoding/base64"
	"path"
	"strconv"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"net/http"

	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"prism/config"
	"prism/database"
	"prism/session"
)

var (
	oidcProvider *oidc.Provider
	oauth2Config oauth2.Config
)

const (
	cookieName          = "session_cookie"
	EmailContextKey     = "email"
	SessionIDContextKey = "sessionID"
)

func init() {
	// Load the configuration
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	oidcProvider, err := oidc.NewProvider(context.Background(), config.AppConfig.OIDC.ProviderURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get provider: %v\n", err)
		os.Exit(1)
	}

	// OAuth2 configuration
	oauth2Config = oauth2.Config{
		ClientID:     config.AppConfig.OIDC.ClientID,
		ClientSecret: config.AppConfig.OIDC.ClientSecret,
		RedirectURL:  config.AppConfig.OIDC.RedirectURI,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

func generateState() string {
	b := make([]byte, 32) // Creates a slice with 32 random bytes
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Failed to generate random state: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b) // Converts bytes to a base64 URL-safe string
}

func RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		userRoleInterface, exists := c.Get("role")
		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		action := actionFromMethod(c.Request.Method)
		resource := ExtractResourcePath(c.Request)

		role, exists := config.AppConfig.Roles[userRoleInterface.(string)]
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			return
		}

		if !hasPermission(role, action, resource) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		} else {
			c.Next()
			return
		}
	}
}

func globalAccess(permissions []config.Permission, path string) bool {
	var globalAccessProject = false
	var globalAccessVulnerability = false
	for _, perm := range permissions {
		if "/project/:id" == perm.Resource {
			globalAccessProject = true
		}
		if "/vulnerability/:id" == perm.Resource {
			globalAccessVulnerability = true
		}
	}

	if path == "/vulnerability" {
		return globalAccessVulnerability
	} else if path == "/project" {
		return globalAccessProject
	}

	return globalAccessProject && globalAccessVulnerability
}

func ACLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("projectID")
		findingsIDStr := c.Param("findingsID")
		email, exists := c.Get("email")
		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		role, _ := c.Get("role")

		if findingsIDStr != "" {

			if globalAccess(config.AppConfig.Roles[role.(string)].Permissions, "/vulnerability") {
				c.Next()
				return
			}

			findingsID, err := strconv.ParseUint(findingsIDStr, 10, 64)
			if err != nil {
				// Handle error if the findingsID is not a valid number
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			projectIDFromVulnerability, err := database.GetProjectIdFromVulnerabilityID(uint(findingsID))
			if err != nil {
				// Handle error from GetProjectIdFromVulnerabilityID
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			// Update the projectID for subsequent checks
			projectID = fmt.Sprintf("%d", projectIDFromVulnerability)
		}

		if projectID != "" {

			if globalAccess(config.AppConfig.Roles[role.(string)].Permissions, "/project") {
				c.Next()
				return
			}

			// Use HasClientAccessToProject to check access
			hasAccess, err := database.HasClientAccessToProject(email.(string), projectID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking project access"})
				c.Abort()
				return
			}

			if !hasAccess {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to the project"})
				c.Abort()
				return
			} else {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusNotFound)
	}
}

func ExtractResourcePath(c *http.Request) string {
	// Definer et sett med forventede startsegmenter
	expectedPrefixes := []string{"/api"}

	// Deler URL-stien basert på '/' og filtrerer ut tomme deler
	parts := strings.Split(c.URL.Path, "/")
	nonEmptyParts := []string{}
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}

	// Sjekk om det første gyldige segmentet matcher et av de forventede startsegmentene
	if len(nonEmptyParts) > 0 && contains(expectedPrefixes, "/"+nonEmptyParts[0]) {
		// Sjekker om det er nok deler til å konstruere ønsket sti
		if len(nonEmptyParts) >= 2 {
			// Returnerer "/api/project" eller tilsvarende basert på den faktiske URL-en
			return "/" + nonEmptyParts[1]
		}
	}

	// Returner en standard eller feil sti hvis stien ikke starter med et forventet segment
	return "/invalid-path" // Eller annen håndtering etter ønske
}

func actionFromMethod(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "write"
	case "PUT":
		return "write"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

func hasPermission(role config.Role, action, resource string) bool {
	denyAccess := false

	for _, perm := range role.Permissions {
		// Global deny if resource is "*" and action list is empty
		if perm.Resource == "*" && len(perm.Action) == 1 {
			if perm.Action[0] == "" {
				return false
			}
		}

		// Specific resource deny
		if perm.Resource == resource && len(perm.Action) == 1 {
			if perm.Action[0] == "" {
				denyAccess = true
				break // No need to check further if access is explicitly denied
			}
		}

		// Check for global access
		if perm.Resource == "*" && contains(perm.Action, action) {
			return true
		}

		// Match specific resource paths
		matched, _ := path.Match(perm.Resource, resource)
		if matched && contains(perm.Action, action) {
			return true
		}
	}

	return denyAccess
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func AuthMiddleware(store *session.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, err := GetSignedCookie(c, cookieName)
		if err != nil {
			// Handle error or invalid session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - no valid cookie"})
			fmt.Println(err)
			return
		}

		validation, err := store.ValidateSession(userInfo.Email, userInfo.SessionID)
		if err != nil {
			// Handle invalid session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - invalid session"})
			return
		}

		if !validation.IsValid {
			// Handle completely invalid session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - session not found"})
			return
		}
		settings, _ := database.GetSettings(false)
		if settings.MFAEnabled == true {
			if !validation.IsOTPVerified && c.Request.URL.Path != "/api/session/otp/generate" && c.Request.URL.Path != "/api/session/otp/validate" {
				// OTP is not verified, initiate OTP verification process
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OTP is not verified", "initiateOTP": true})
				return
			}
		}

		// Add email to Gin context for easier access in subsequent handlers
		c.Set(EmailContextKey, userInfo.Email)
		c.Set(SessionIDContextKey, userInfo.SessionID)

		role := store.GetRole(userInfo.Email)
		permissions := config.AppConfig.Roles[role].Permissions

		c.Set("isGlobalProject", globalAccess(permissions, "/project"))
		c.Set("isGlobalVulnerability", globalAccess(permissions, "/vulnerability"))
		c.Set("role", store.GetRole(userInfo.Email))

		c.Next()
	}
}

func HandleLogin(c *gin.Context) {
	// Redirect to the OIDC provider's login page
	state := generateState()

	setSecureStateCookie(c, state)

	c.Redirect(http.StatusTemporaryRedirect, oauth2Config.AuthCodeURL(state))
}

type UserInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	SessionID string `json:"sessionId"`
}

func HandleUserRequest(c *gin.Context, store *session.SessionStore) {
	userInfo, err := GetSignedCookie(c, cookieName)
	if err != nil {
		// Handle error or invalid session
		c.AbortWithStatus(http.StatusUnauthorized)
		fmt.Println(err)
		return
	}

	user, _ := store.GetUserDataByEmail(userInfo.Email)

	c.JSON(http.StatusOK, user)
}

func HandleLogout(c *gin.Context, store *session.SessionStore) {
	// Retrieve email from Gin context
	emailInterface, exists := c.Get(EmailContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user email not found in context"})
		return
	}
	email, ok := emailInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid email format in context"})
		return
	}

	ClearSignedCookie(c, cookieName)

	// Invalidate the session
	if err := store.InvalidateSession(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to invalidate session"})
		return
	}

	c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin)
}

func HandleCallback(c *gin.Context, store *session.SessionStore) {

	receivedState, err := c.Cookie("oidc_state")
	if err != nil {
		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("OIDC Cookie state is not found. Contact administrator. You could be a victim of a CSRF attack!"))
		return
	}

	// Compare with the state in the query parameter
	queryState := c.Query("state")
	if receivedState != queryState {
		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("Cookie state does not match. Contact administrator. You could be a victim of a CSRF attack!"))
		return
	}

	// Extract code and state from query parameters
	code := c.Query("code")

	// Exchange the code for a token
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Extract the ID token from OAuth2 token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token field in OAuth2 token"})
		return
	}

	// Parse the token without validating the signature
	tokenClaims, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		// Handle error
		fmt.Println("Error parsing token:", err)
		return
	}

	if claims, ok := tokenClaims.Claims.(jwt.MapClaims); ok {
		userInfo := UserInfo{
			Name:    getStringFromMapClaims(claims, "name"),
			Email:   getStringFromMapClaims(claims, "email"),
			Picture: getStringFromMapClaims(claims, "picture"),
		}

		userInfo.SessionID = uuid.New().String()

		database.SaveOrUpdateUserData(userInfo.Name, userInfo.Email, userInfo.Picture)

		store.PersistSession(userInfo.Email, userInfo.SessionID)

		SetSignedCookie(c, cookieName, userInfo)

		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin)
	}

}

func getStringFromMapClaims(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func setSecureStateCookie(c *gin.Context, state string) {
	expirationTime := time.Now().Add(5 * time.Minute)

	cookie := &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Expires:  expirationTime,
		Path:     "/api/callback",
		Domain:   "", // Current domain
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Setting SameSite to Strict
	}

	http.SetCookie(c.Writer, cookie)
}
