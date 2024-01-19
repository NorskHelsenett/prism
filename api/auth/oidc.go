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
		emailInterface, exists := c.Get(EmailContextKey)
		if !exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "user email not found in context"})
				return
		}

		email, ok := emailInterface.(string)
		if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "invalid email format in context"})
				return
		}

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

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Access forbidden",
			})
			return
		}
		c.Next()
	}
}

func AuthMiddleware(store *session.SessionStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        cookie, err := GetSignedCookie(c, cookieName)
        if err != nil {
            // Handle error or invalid session
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - no valid cookie"})
            fmt.Println(err)
            return
        }

        userInfo, err := DecodeCookieAndUnmarshal(c, cookie)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            fmt.Println(err)
            return
        }

        validation, err := store.ValidateSession(userInfo.Email)
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

        if !validation.IsOTPVerified && c.Request.URL.Path != "/api/session/otp/generate"  && c.Request.URL.Path != "/api/session/otp/validate" && c.Request.URL.Path != "/api/session/otp/reset" {
            // OTP is not verified, initiate OTP verification process
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OTP is not verified", "initiateOTP": true})
            return
        }

        // Add email to Gin context for easier access in subsequent handlers
        c.Set(EmailContextKey, userInfo.Email)

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

	user, _ := database.GetUserDataByEmail(userInfo.Email)

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

	appConfig, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	c.Redirect(http.StatusFound, appConfig.Cors.Origin)
}

func HandleCallback(c *gin.Context, store *session.SessionStore) {
	appConfig, _ := config.LoadConfig()

	receivedState, err := c.Cookie("oidc_state")
	if err != nil {
		c.Redirect(http.StatusFound, appConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("OIDC Cookie state is not found. Contact administrator. You could be a victim of a CSRF attack!"))
		return
	}

	// Compare with the state in the query parameter
	queryState := c.Query("state")
	if receivedState != queryState {
		c.Redirect(http.StatusFound, appConfig.Cors.Origin+"/error.html?message="+url.QueryEscape("Cookie state does not match. Contact administrator. You could be a victim of a CSRF attack!"))
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

		store.PersistSession(userInfo.Email)

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
