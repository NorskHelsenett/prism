package auth

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"github.com/golang-jwt/jwt"

	"net/http"

	"context"
	"os"
    "fmt"
    "encoding/json"

    "prism/database"
	"prism/config"
)

var (
    oidcProvider *oidc.Provider
    oauth2Config oauth2.Config
)

const (
    cookieName = "session_cookie"
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
	ClientID: appConfig.OIDC.ClientID,
	ClientSecret: appConfig.OIDC.ClientSecret,
	RedirectURL: appConfig.OIDC.RedirectURI,
        Endpoint:     oidcProvider.Endpoint(),
        Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
    }
}

func generateState() string {
    b := make([]byte, 32) // Creates a slice with 32 random bytes
    _, err := rand.Read(b)
    if err != nil {
        // Handle error
    }
    return base64.URLEncoding.EncodeToString(b) // Converts bytes to a base64 URL-safe string
}

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        email,_ := c.Request.Context().Value(EmailContextKey).(string)

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
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        c.Next()
    }
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
    cookie, err := getSignedCookie(c, cookieName)
    if err != nil {
        // Handle error or invalid session
        c.AbortWithStatus(http.StatusUnauthorized)
        fmt.Println(err)
        return
    }

    userInfo, err := decodeCookieAndUnmarshal(c, cookie)
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
    c.Redirect(http.StatusTemporaryRedirect, oauth2Config.AuthCodeURL(state))
}

type UserInfo struct {
    Name    string `json:"name"`
    Email   string `json:"email"`
    Picture string `json:"picture"`
}

// Extracted function to decode the cookie and unmarshal JSON
func decodeCookieAndUnmarshal(c *gin.Context, cookie string) (UserInfo, error) {
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
    cookie, err := getSignedCookie(c, cookieName)
    if err != nil {
        // Handle error or invalid session
        c.AbortWithStatus(http.StatusUnauthorized)
        fmt.Println(err)
        return
    }

    userInfo, err := decodeCookieAndUnmarshal(c, cookie)
    if err != nil {
        c.AbortWithStatus(http.StatusInternalServerError)
        return
    }

    c.JSON(http.StatusOK, userInfo)
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
        setSignedCookie(c, cookieName, encodedJSON)
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
