package messages

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"time"
)

type Message struct {
	ServerId   string
	TargetId   string
	SentAt     string
	SentAtTime time.Time
	MessageId  gocql.UUID
	Content    string
	Edited     bool
	Reactions  map[gocql.UUID]string
	SentBy     gocql.UUID
}

func LoadMessages(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(jwt, session); ret.Status {
			var messages []Message
			var body struct {
				Target gocql.UUID `json:"target"`
				Type   string     `json:"type"`
			}
			if err := c.BindJSON(&body); err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			if body.Type == "DM" {
				target := CombineUUIDs(ret.User.UserId, body.Target)
				iter := session.Query(`SELECT * FROM messages WHERE target_id = ? AND server_id = ? LIMIT 50`, target, "").Iter()
				for {
					var msg Message

					// Scan the current row into msg
					if !iter.Scan(&msg.ServerId, &msg.TargetId, &msg.SentAt, &msg.SentAtTime, &msg.MessageId, &msg.Content, &msg.Edited, &msg.Reactions, &msg.SentBy) {
						break
					}
					messages = append(messages, msg)
				}
				c.JSON(http.StatusOK, messages)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
			return
		}
	}
}
