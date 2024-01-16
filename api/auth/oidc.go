package auth

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"

	"net/http"

	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"net/url"
	"log"

	"prism/config"
	"prism/database"
)

var (
	oidcProvider *oidc.Provider
	oauth2Config oauth2.Config
)

const (
	cookieName      = "session_cookie"
	EmailContextKey = "email"
)

func init() {
	// Setup the OIDC provider
	appConfig, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	oidcProvider, err = oidc.NewProvider(context.Background(), appConfig.OIDC.ProviderURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get provider: %v\n", err)
		os.Exit(1)
	}

	// OAuth2 configuration
	oauth2Config = oauth2.Config{
		ClientID:     appConfig.OIDC.ClientID,
		ClientSecret: appConfig.OIDC.ClientSecret,
		RedirectURL:  appConfig.OIDC.RedirectURI,
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

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, _ := c.Request.Context().Value(EmailContextKey).(string)

		appConfig, _ := config.LoadConfig()

		admins := appConfig.Admins

		// Check if email is in admins
		isAdmin := false
		for _, adminEmail := range admins {
			if email == adminEmail {
				isAdmin = true
				break
			}
		}
		// 403 since we are testing for authenticated users before running this middleware.
		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Access forbidden",
			})
			return
		}
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := GetSignedCookie(c, cookieName)
		if err != nil {
			// Handle error or invalid session
			c.AbortWithStatus(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}

		userInfo, err := DecodeCookieAndUnmarshal(c, cookie)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(c.Request.Context(), EmailContextKey, userInfo.Email)
		c.Request = c.Request.WithContext(ctx)

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
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

// Extracted function to decode the cookie and unmarshal JSON
func DecodeCookieAndUnmarshal(c *gin.Context, cookie string) (UserInfo, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(cookie)
	if err != nil {
		// handle base64 decoding error
		fmt.Println(err)
		return UserInfo{}, err
	}

	decodedValue := string(decodedBytes)

	var userInfo UserInfo
	err = json.Unmarshal([]byte(decodedValue), &userInfo)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return UserInfo{}, err
	}

	return userInfo, nil
}

func HandleUserRequest(c *gin.Context) {
	cookie, err := GetSignedCookie(c, cookieName)
	if err != nil {
		// Handle error or invalid session
		c.AbortWithStatus(http.StatusUnauthorized)
		fmt.Println(err)
		return
	}

	userInfo, err := DecodeCookieAndUnmarshal(c, cookie)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user,_ := database.GetUserDataByEmail(userInfo.Email)

	c.JSON(http.StatusOK, user)
}

func HandleLogout(c *gin.Context) {
	ClearSignedCookie(c, cookieName)
	appConfig, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	c.Redirect(http.StatusFound, appConfig.Cors.Origin+"/login")
}

func HandleCallback(c *gin.Context) {
	appConfig, _ := config.LoadConfig()

	receivedState, err := c.Cookie("oidc_state")
	if err != nil {
			c.Redirect(http.StatusFound, appConfig.Cors.Origin + "/error.html?message="+url.QueryEscape("OIDC Cookie state is not found. Contact administrator. You could be a victim of a CSRF attack!"))
			return
	}

	// Compare with the state in the query parameter
	queryState := c.Query("state")
	if receivedState != queryState {
			c.Redirect(http.StatusFound, appConfig.Cors.Origin +"/error.html?message="+url.QueryEscape("Cookie state does not match. Contact administrator. You could be a victim of a CSRF attack!"))
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

		database.SaveOrUpdateUserData(userInfo.Name, userInfo.Email, userInfo.Picture)

		jsonValue, err := json.Marshal(userInfo)
		if err != nil {
			// handle JSON marshaling error
			return
		}

		// Base64 encode the JSON string
		encodedJSON := base64.StdEncoding.EncodeToString(jsonValue)
		SetSignedCookie(c, cookieName, encodedJSON)
		appConfig, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		c.Redirect(http.StatusFound, appConfig.Cors.Origin)
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
    expirationSeconds := int(expirationTime.Sub(time.Now()).Seconds())

    c.SetCookie(
        "oidc_state",      // Name of the cookie
        state,             // Value of the cookie (state string)
        expirationSeconds, // Max-Age of the cookie in seconds (10 minutes)
        "/api/callback",  // Path for which the cookie is valid
        "",                // Domain for which the cookie is valid (empty string means current domain)
        true,              // Secure flag (true means send only over HTTPS)
        true,              // HttpOnly flag (true means the cookie is not accessible via JavaScript)
    )
}