package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/securecookie"
    "net/http"
    "time"
    "net/url"
    "prism/config"
    "os"
)

var hashKey = securecookie.GenerateRandomKey(64)
var blockKey = securecookie.GenerateRandomKey(32)
var secure = securecookie.New(hashKey, blockKey)

func SetSignedCookie(c *gin.Context, name, value string) {

    // Encode the base64 string using securecookie
    encoded, err := secure.Encode(name, value)
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
        Name:     name,
        Value:    encoded,
        Path:     "/",
        Domain:   domain,
        // SameSite: http.SameSiteStrictMode,
        HttpOnly: true, // Recommended
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

func GetSignedCookie(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
    if err != nil {
        return "", err
    }

    var value string
    err = secure.Decode(name, cookie.Value, &value)
    if err != nil {
        return "", err
    }

    return value, nil
}

func ClearSignedCookie(c *gin.Context, name string) {
    // Set a cookie with the same name, but with an expiration date in the past
    // c.SetCookie(name, "", -1, "/", "", false, true)
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			HttpOnly: true, // Recommended
			Secure:   false, // Set to true if using HTTPS, Required when SameSite=None
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		}

    http.SetCookie(c.Writer, cookie)
}
