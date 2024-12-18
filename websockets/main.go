package websockets

import (
	"backend-joltamp/analytics"
	"backend-joltamp/security"
	"backend-joltamp/types"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebSocket upgrader with buffer sizes and handshake timeout
var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/* USER ID -> WS Conn */
var ConnactedUsers = make(map[gocql.UUID]*websocket.Conn)
var mu sync.Mutex

// WebsocketHandler handles incoming WebSocket connection requests
func WebsocketHandler(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set WebSocket upgrade:", err)
			return
		}
		defer wsConn.Close()

		// Parse JWT from the Authorization header
		jwt, err := gocql.ParseUUID(c.Query("token"))
		if err != nil {
			wsConn.WriteJSON(gin.H{
				"type":    "disconnacted",
				"payload": "JWT cant be verified!",
			})
			return
		}
		ret := security.VerifyJWT(jwt.String(), session)
		if !ret.Status {
			wsConn.WriteJSON(gin.H{
				"type":    "disconnacted",
				"payload": "Wrong JWT token",
			})
			return
		}
		// Add user connection to the map
		mu.Lock()
		ConnactedUsers[ret.User.UserId] = wsConn
		mu.Unlock()

		// Log and notify the connection success
		fmt.Println("User connacted to socket")
		analytics.OnWebsocketsConnect()
		wsConn.WriteJSON(gin.H{
			"type":    "connacted",
			"payload": "Connacted to ws",
		})
		fmt.Println(len(ConnactedUsers))

		// Start a heartbeat ticker to keep the connection alive
		func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				<-ticker.C
				err := wsConn.WriteJSON(gin.H{
					"type":    "heartbeat",
					"payload": nil,
				})
				if err != nil {
					fmt.Println("User disconnacted from socket")
					delete(ConnactedUsers, ret.User.UserId)
					return
				}
			}
		}()
	}
}

// HandleMessageSendWS sends a new message to the target user via WebSocket
func HandleMessageSendWS(server string, target string, message types.Message) {
	if server == "" {
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists {
			wsConn.WriteJSON(gin.H{
				"type":    "new_message",
				"payload": message,
			})
		}
	}
}

// HandleMessageDeleteWS notifies the target user of a deleted message
func HandleMessageDeleteWS(server string, target string, message types.DeleteMessage) {
	if server == "" {
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists {
			wsConn.WriteJSON(gin.H{
				"type":    "delete_message",
				"payload": message,
			})
		}
	}
}

// HandleMessageEditWS notifies the target user of an edited message
func HandleMessageEditWS(server string, target string, message types.EditMessage) {
	if server == "" {
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists {
			wsConn.WriteJSON(gin.H{
				"type":    "edit_message",
				"payload": message,
			})
		}
	}
}
