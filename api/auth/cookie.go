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

	domain, err := getDomainFromURL(config.AppConfig.Cors.Origin)
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
		MaxAge:   3600 * 8,
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

func SetSignedCookieFor(c *gin.Context, cookieName, path, value string, maxAge int, laxMode ...bool) {

	laxModeValue := false
	if len(laxMode) > 0 {
		laxModeValue = laxMode[0]
	}

	sameSiteMode := http.SameSiteStrictMode
	if laxModeValue {
		sameSiteMode = http.SameSiteLaxMode
	}

	encoded, err := secure.Encode(cookieName, value)
	if err != nil {
		return
	}

	secureFlag := true
	if os.Getenv("GO_ENV") == "dev" {
		secureFlag = false
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     path,
		Domain:   "",
		SameSite: sameSiteMode,
		HttpOnly: true,
		Secure:   secureFlag,
		MaxAge:   maxAge,
	}

	http.SetCookie(c.Writer, cookie)
}
func GetSignedCookieFor(c *gin.Context, cookieName string) (string, error) {
	cookie, err := c.Request.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	var value string
	err = secure.Decode(cookieName, cookie.Value, &value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func ClearSignedCookie(c *gin.Context, name string) {
	domain, _ := getDomainFromURL(config.AppConfig.Cors.Origin)

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
