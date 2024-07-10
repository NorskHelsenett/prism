package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	rateLimit    = 5 // requests per minute
	cleanupTime  = 10 * time.Minute
	requestStore = make(map[string]*requestCounter)
	mu           sync.Mutex
)

type requestCounter struct {
	count     int
	timestamp time.Time
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		mu.Lock()
		if _, exists := requestStore[clientIP]; !exists {
			requestStore[clientIP] = &requestCounter{timestamp: time.Now()}
		}

		reqInfo := requestStore[clientIP]
		if time.Since(reqInfo.timestamp) > time.Minute {
			reqInfo.count = 1
			reqInfo.timestamp = time.Now()
		} else {
			reqInfo.count++
		}

		if reqInfo.count > rateLimit {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		mu.Unlock()

		c.Next()
	}
}

func init() {
	go func() {
		for {
			time.Sleep(cleanupTime)
			mu.Lock()
			for key, val := range requestStore {
				if time.Since(val.timestamp) > cleanupTime {
					delete(requestStore, key)
				}
			}
			mu.Unlock()
		}
	}()
}
