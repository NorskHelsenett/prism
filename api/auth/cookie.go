package auth

import (
	"net/http"
	"net/url"
	"os"
	"prism/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

var hashKey = securecookie.GenerateRandomKey(64)
var blockKey = securecookie.GenerateRandomKey(32)
var secure = securecookie.New(hashKey, blockKey)

func SetSignedCookie(c *gin.Context, cookieName string, userInfo UserInfo) {

	// Encode the base64 string using securecookie
	encoded, err := secure.Encode(cookieName, userInfo)
	if err != nil {
		// handle secure encoding error
		return
	}

	appConfig, _ := config.LoadConfig()
	domain, err := getDomainFromURL(appConfig.Cors.Origin)
	if err != nil {
		// handle error
	}

	secureFlag := true
	if os.Getenv("GO_ENV") == "dev" {
		secureFlag = false
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,       // Recommended
		Secure:   secureFlag, // Set to true if using HTTPS, Required when SameSite=None
		MaxAge:   3600 * 12,
	}

	http.SetCookie(c.Writer, cookie)
}

func getDomainFromURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

func GetSignedCookie(c *gin.Context, name string) (UserInfo, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return UserInfo{}, err
	}

	// Define a structure to hold the decoded values
	var userInfo UserInfo

	err = secure.Decode(cookieName, cookie.Value, &userInfo)
	if err != nil {
		return UserInfo{}, err
	}

	return userInfo, nil
}

func ClearSignedCookie(c *gin.Context, name string) {
	appConfig, _ := config.LoadConfig()
	domain, _ := getDomainFromURL(appConfig.Cors.Origin)

	// Set a cookie with the same name, but with an expiration date in the past
	// c.SetCookie(name, "", -1, "/", "", false, true)
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}

	http.SetCookie(c.Writer, cookie)
}
