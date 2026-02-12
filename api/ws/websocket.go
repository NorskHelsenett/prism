package ws

import (
	"fmt"
	"net/http"
	"prism/auth"
	"prism/database"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func WSHandler(c *gin.Context) {

	cookie, err := auth.GetSignedCookie(c, "session_cookie")
	if err != nil || cookie.Email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	go func() {
		shouldReturn := sendWebsocketMessageFor(cookie, conn)
		if shouldReturn {
			return
		}
		for range ticker.C {
			shouldReturn := sendWebsocketMessageFor(cookie, conn)
			if shouldReturn {
				return
			}
		}
	}()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			break
		}
	}
}

func sendWebsocketMessageFor(cookie auth.UserInfo, conn *websocket.Conn) bool {
	notifications, err := database.GetNotifications(cookie.Email)
	if err != nil {
		fmt.Printf("Error getting notifications: %v\n", err)
		return true
	}
	msg := WebsocketMessage{
		Type: "notifications",
		Data: notifications,
	}
	if err := conn.WriteJSON(msg); err != nil {
		fmt.Println("Write error:", err)
		return true
	}
	return false
}
