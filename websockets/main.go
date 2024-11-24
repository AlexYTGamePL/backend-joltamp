package websockets

import (
	"backend-joltamp/security"
	"backend-joltamp/types"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"fmt"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	HandshakeTimeout: 5000,
}

/* USER ID -> WS Conn */
var ConnactedUsers = make(map[gocql.UUID]*websocket.Conn)
var mu sync.Mutex

func WebsocketHandler(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set WebSocket upgrade:", err)
			return
		}
		defer wsConn.Close()
		jwt, err := gocql.ParseUUID(c.GetHeader("Authorization"))
		if err != nil{
			wsConn.WriteJSON(gin.H{
				"type": "disconnacted",
				"payload": "JWT cant be verified!",
			})
			return
		}
		ret := security.VerifyJWT(jwt.String(), session);
		if !ret.Status{
			wsConn.WriteJSON(gin.H{
				"type": "disconnacted",
				"payload": "Wrong JWT token",
			})
			return
		}
		mu.Lock()
		ConnactedUsers[ret.User.UserId] = wsConn
		mu.Unlock()
		fmt.Println("User connacted to socket")
		wsConn.WriteJSON(gin.H{
			"type": "connacted",
			"payload": "Connacted to ws",
		})
		fmt.Println(len(ConnactedUsers))
		func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				<- ticker.C
				err := wsConn.WriteJSON(gin.H{
					"type": "heartbeat",
					"payload": nil,
				})
				if err != nil{
					fmt.Println("User disconnacted from socket")
					delete(ConnactedUsers, ret.User.UserId)
					return
				}
			}
		}()
	}
}

func HandleMessageSendWS(server string, target string, message types.Message){
	if server == ""{
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists{
			wsConn.WriteJSON(gin.H{
				"type": "new_message",
				"payload": message,
			})
		}
	}
}

func HandleMessageDeleteWS(server string, target string, message gocql.UUID){
	if server == ""{
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists{
			wsConn.WriteJSON(gin.H{
				"type": "delete_message",
				"payload": message,
			})
		}
	}
}


func HandleMessageEditWS(server string, target string, message types.EditMessage){
	if server == ""{
		targetUUID, _ := gocql.ParseUUID(target)
		if wsConn, exists := ConnactedUsers[targetUUID]; exists{
			wsConn.WriteJSON(gin.H{
				"type": "edit_message",
				"payload": message,
			})
		}
	}
}