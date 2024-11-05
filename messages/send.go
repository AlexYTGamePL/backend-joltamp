package messages

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func SendMessage(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(jwt, session); ret.Status {

			var body struct {
				Target  gocql.UUID  `json:"target"`
				Server  *gocql.UUID `json:"server"`
				Content string      `json:"content"`
				Reply   *gocql.UUID `json:"reply"`
			}
			if err := c.BindJSON(&body); err != nil {
				c.Status(http.StatusBadRequest)
				return
			}

			// Checking if user isnt trying to send message to yourself
			if body.Target == ret.User.UserId{
				c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "You cant message yourself"})
				return
			}

			// Checking if user is sending message on dm/server
			var target string
			if body.Server == nil {
				target = CombineUUIDs(ret.User.UserId, body.Target)
			} else {
				target = body.Target.String()
			}

			messageId, _ := gocql.RandomUUID()

			// Sending a message with self-running functions
			if err := session.Query(`INSERT INTO messages (server_id, target_id, sent_at, sent_at_time, message_id, content, edited, reactions, reply, sent_by) VALUES (?, ?, toDate(now()), toTimeStamp(now()), ?, ?, false, null, ?, ?)`, func() string {
				if body.Server != nil {
					return body.Server.String()
				} else {
					return ""
				}
			}(), target, messageId, body.Content, func() string {
				if body.Reply != nil {
					return body.Reply.String()
				} else {
					return ""
				}
			}(), ret.User.UserId).Exec(); err != nil {
				println("47 - " + err.Error())
				c.Status(http.StatusInternalServerError)
				return
			}

			// Sending back message
			var insertedMessage struct {
				ServerId   string
				TargetId   string
				SentAt     string
				SentAtTime int64
				MessageId  gocql.UUID
				Content    string
				Edited     bool
				Reactions  map[gocql.UUID]string
				Reply      string
				SentBy     gocql.UUID
			}

			insertedMessageId := messageId
			if err := session.Query(`SELECT * FROM messages WHERE server_id = '' AND target_id = ? AND message_id = ? ALLOW FILTERING`, target, insertedMessageId).Scan(
				&insertedMessage.ServerId,
				&insertedMessage.TargetId,
				&insertedMessage.SentAt,
				&insertedMessage.SentAtTime,
				&insertedMessage.MessageId,
				&insertedMessage.Content,
				&insertedMessage.Edited,
				&insertedMessage.Reactions,
				&insertedMessage.Reply,
				&insertedMessage.SentBy,
			); err != nil {
				c.Status(http.StatusInternalServerError)
				println(err.Error())
				return
			}

			c.JSON(200, insertedMessage)
		}else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
			return
		}
	}
}
