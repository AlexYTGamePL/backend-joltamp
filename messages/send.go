package messages

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"time"
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
			var target string
			if body.Server == nil {
				target = CombineUUIDs(ret.User.UserId, body.Target)
			} else {
				target = body.Target.String()
			}

			messageId, _ := gocql.RandomUUID()
			timeVar := time.Now().UTC()
			if err := session.Query(`INSERT INTO messages (server_id, target_id, sent_at, sent_at_time, message_id, content, edited, reactions, reply, sent_by) VALUES (?, ?, toDate(now()), ?, ?, ?, false, null, ?, ?)`, func() string {
				if body.Server != nil {
					return body.Server.String()
				} else {
					return ""
				}
			}(), target, timeVar, messageId, body.Content, func() string {
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

			var insertedMessage struct {
				ServerId   string
				TargetId   string
				SentAt     string
				SentAtTime time.Time
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
		}
	}
}
