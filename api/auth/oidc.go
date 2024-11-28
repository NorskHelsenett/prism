package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"

	"os"
	"path"
	"strconv"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"crypto/sha256"
	"net/http"

	"context"
	"fmt"
	"log"
	"net/url"

	"prism/config"
	"prism/database"
	"prism/routes"
	"prism/session"
)

var (
	oidcProvider *oidc.Provider
	oauth2Config map[string]*oauth2.Config
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
	//it was here
	oauth2Config = make(map[string]*oauth2.Config)

	for name, oidcConfig := range config.AppConfig.OIDC {
		oidcProvider, err := oidc.NewProvider(context.Background(), oidcConfig.ProviderURI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get provider: %v\n", err)
			os.Exit(1)
		}

		oauth2Config[name] = &oauth2.Config{
			ClientID:     oidcConfig.ClientID,
			ClientSecret: oidcConfig.ClientSecret,
			RedirectURL:  oidcConfig.RedirectURI,
			Endpoint:     oidcProvider.Endpoint(),
			// You might need to specify Scopes depending on your provider
			Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
		}
	}
}

func generateState(provider string) string {
	b := make([]byte, 32) // Creates a slice with 32 random bytes
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Failed to generate random state: %v", err)
	}
	stateToken := base64.URLEncoding.EncodeToString(b) // Converts bytes to a base64 URL-safe string
	return stateToken + ":" + provider
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
			//@todo should this be soft-reset to visitor?
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

			vulnerability, err := database.GetVulnerabilityIds(false, email.(string), []uint{uint(findingsID)})
			if err != nil {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			hasAccessToVulnerability := len(vulnerability) > 0

			if hasAccessToVulnerability {
				c.Next()
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

		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			if email, valid := routes.ValidateAPIKey(apiKey); valid {
				c.Set(EmailContextKey, email)

				role := store.GetRole(email)
				permissions := config.AppConfig.Roles[role].Permissions

				c.Set("isGlobalProject", globalAccess(permissions, "/project"))
				c.Set("isGlobalVulnerability", globalAccess(permissions, "/vulnerability"))
				c.Set("role", role)

				c.Next()
				return
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		}

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
		if settings.MFAEnabled {
			if !validation.IsOTPVerified && !strings.HasPrefix(c.Request.URL.Path, "/api/share/") && c.Request.URL.Path != "/api/session/otp/generate" && c.Request.URL.Path != "/api/session/otp/validate" {
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
		c.Set("role", role)

		c.Next()
	}
}

func generateCodeVerifier() string {
	b := make([]byte, 32) // 32 bytes give 256 bits entropy
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Beregn Code Challenge basert på Code Verifier
func generateCodeChallenge(codeVerifier string) string {
	s256 := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(s256[:])
}

func HandleLogin(c *gin.Context) {
	provider := c.Query("provider")

	if oauthConfig, ok := oauth2Config[provider]; ok {
		// Redirect to the OIDC provider's login page
		state := generateState(provider)
		SetSignedCookieFor(c, "oidc_state", "/api/callback", state, 69, true)

		codeVerifier := generateCodeVerifier()
		if codeVerifier == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to complete request"})
		}

		codeChallenge := generateCodeChallenge(codeVerifier)
		SetSignedCookieFor(c, "code_verifier", "/api/callback", codeVerifier, 69, true)
		authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("code_challenge", codeChallenge), oauth2.SetAuthURLParam("code_challenge_method", "S256"))

		c.Header("Referrer-Policy", "no-referrer")
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	} else {
		// Handle error: provider not configured
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
	}
}

type UserInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
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

	receivedState, err := GetSignedCookieFor(c, "oidc_state")
	if err != nil {
		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("OIDC Cookie state is not found. Contact administrator. You could be a victim of a CSRF attack!"))
		return
	}

	// Hent Code Verifier fra cookie eller annen lagring
	codeVerifier, _ := GetSignedCookieFor(c, "code_verifier")

	// Compare with the state in the query parameter
	queryState := c.Query("state")
	if receivedState != queryState {
		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("Cookie state does not match. Contact administrator. You could be a victim of a CSRF attack!"))
		return
	}
	stateParts := strings.SplitN(queryState, ":", 2)

	// Now you know the provider
	provider := stateParts[1]
	oauthConfig, _ := oauth2Config[provider]

	// Extract code and state from query parameters
	code := c.Query("code")

	// Exchange the code for a token
	token, err := oauthConfig.Exchange(context.Background(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Extract the ID token from OAuth2 token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token field in OAuth2 token"})
		return
	}

	// Parse and verify the token
	parser := &jwt.Parser{
		ValidMethods: []string{"RS256", "RS384", "RS512"}, // Allow only RS256 signing method
	}

	claims := jwt.MapClaims{}

	// Parse with validation
	tokenClaims, err := parser.ParseWithClaims(idToken, claims, keyFunc(provider))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
		return
	}

	if !tokenClaims.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signature"})
		return
	}
	profilePicture := ""
	if claims, ok := tokenClaims.Claims.(jwt.MapClaims); ok {
		profilePicture = getStringFromMapClaims(claims, "picture")
		userInfo := UserInfo{
			Name:  getStringFromMapClaims(claims, "name"),
			Email: strings.ToLower(getStringFromMapClaims(claims, "email")),
		}

		userInfo.SessionID = uuid.New().String()

		if provider == "azure" {
			profilePictureAzure, err := getAzureProfilePicture(userInfo.Email, token)
			if err != nil {
				log.Printf("Error getting azure profile picture %s", err)
			}
			if profilePictureAzure != "" {
				profilePicture = profilePictureAzure
			}
		}

		database.SaveOrUpdateUserData(userInfo.Name, userInfo.Email, profilePicture)
		store.SaveOrUpdateUserData(userInfo.Name, userInfo.Email, profilePicture)

		store.PersistSession(userInfo.Email, userInfo.SessionID)

		SetSignedCookie(c, cookieName, userInfo)

		c.Redirect(http.StatusFound, config.AppConfig.Cors.Origin)
	}
}

// JWK represents a JSON Web Key
type JWK struct {
	KID string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSResponse represents the JWKS response structure
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type OpenIDConfiguration struct {
	Jwks_uri string `json:"jwks_uri"`
}

// keyFunc creates a jwt.Keyfunc that uses the JWKS to verify signatures
func keyFunc(provider string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found in token")
		}

		// Get provider configuration
		providerConfig := config.AppConfig.OIDC[provider]

		resp, _ := http.Get(providerConfig.ProviderURI + "/.well-known/openid-configuration")
		var openidConfig OpenIDConfiguration
		if err := json.NewDecoder(resp.Body).Decode(&openidConfig); err != nil {
			return nil, fmt.Errorf("YOU DIED %v", err)
		}
		defer resp.Body.Close()

		// Fetch JWKS from the provider
		resp, err := http.Get(openidConfig.Jwks_uri)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
		}
		defer resp.Body.Close()

		var jwks JWKSResponse
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			return nil, fmt.Errorf("failed to decode JWKS: %v", err)
		}

		// Find the key matching the kid
		for _, key := range jwks.Keys {
			if key.KID == kid {
				return constructRSAPublicKey(key.N, key.E)
			}
		}

		return nil, fmt.Errorf("no matching key found for kid: %s", kid)
	}
}

func decodeBase64(s string) ([]byte, error) {
	// Add padding if needed
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

// constructRSAPublicKey creates an RSA public key from modulus and exponent
func constructRSAPublicKey(n, e string) (*rsa.PublicKey, error) {
	// Decode the modulus
	modulus, err := decodeBase64(n)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %v", err)
	}

	// Decode the exponent
	exponent, err := decodeBase64(e)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %v", err)
	}

	// Convert modulus bytes to big int
	n_big := new(big.Int).SetBytes(modulus)

	// Convert exponent bytes to int
	var e_int int
	for i := 0; i < len(exponent); i++ {
		e_int = e_int*256 + int(exponent[i])
	}

	return &rsa.PublicKey{
		N: n_big,
		E: e_int,
	}, nil
}

func getAzureProfilePicture(email string, token *oauth2.Token) (string, error) {
	accessToken, ok := token.Extra("access_token").(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("no access_token field in OAuth2 token")
	}

	photoURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/photos/96x96/$value", email)

	photoReq, err := http.NewRequest("GET", photoURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for photo: %s", err)
	}
	photoReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	photoResp, err := client.Do(photoReq)
	if err != nil {
		return "", fmt.Errorf("error retrieving photo: %s", err)
	}
	defer photoResp.Body.Close()

	if photoResp.StatusCode == http.StatusOK {
		photoData, err := io.ReadAll(photoResp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading photo data: %s", err)
		}

		contentType := photoResp.Header.Get("Content-Type")
		base64Photo := base64.StdEncoding.EncodeToString(photoData)
		return fmt.Sprintf("data:%s;base64,%s", contentType, base64Photo), nil
	}

	// Optionally read the response body for a more informative error message
	bodyBytes, _ := io.ReadAll(photoResp.Body)
	bodyString := string(bodyBytes)

	return "", fmt.Errorf("failed to get user photo. Status Code: %d. Response: %s", photoResp.StatusCode, bodyString)
}

func getStringFromMapClaims(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
