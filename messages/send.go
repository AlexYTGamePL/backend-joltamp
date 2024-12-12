package messages

import (
	"backend-joltamp/analytics"
	"backend-joltamp/security"
	"backend-joltamp/types"
	"backend-joltamp/websockets"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type ReplyBodyType_Send struct{
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
			target := func() string {
				if body.Server != nil {
					return body.Server.String()
				} else {
					return CombineUUIDs(ret.User.UserId, body.Target)
				}
			}()
			server := func() string {
				if body.Server != nil {
					return body.Server.String()
				} else {
					return ""
				}
			}()

			messageId, _ := gocql.RandomUUID()

			// Sending a message with self-running functions
			if err := session.Query(`INSERT INTO messages (server_id, target_id, sent_at, sent_at_time, message_id, content, edited, reactions, reply, sent_by) VALUES (?, ?, toDate(now()), toTimeStamp(now()), ?, ?, false, null, ?, ?)`, server, target, messageId, body.Content, func() string {
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
			var insertedMessage types.Message

			insertedMessageId := messageId
			if err := session.Query(`SELECT * FROM messages WHERE server_id = ? AND target_id = ? AND message_id = ? ALLOW FILTERING`, server, target, insertedMessageId).Scan(
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

			if insertedMessage.Reply != ""{
				insertedMessage.ReplyBody = &types.ReplyBodyType{}
				if err := session.Query(`SELECT server_id, target_id, sent_at, sent_at_time, message_id, content, edited, reactions, sent_by FROM messages WHERE server_id = ? AND target_id = ? AND message_id = ? ALLOW FILTERING`, server, target, insertedMessage.Reply).Scan(
					&insertedMessage.ReplyBody.ServerId,
					&insertedMessage.ReplyBody.TargetId,
					&insertedMessage.ReplyBody.SentAt,
					&insertedMessage.ReplyBody.SentAtTime,
					&insertedMessage.ReplyBody.MessageId,
					&insertedMessage.ReplyBody.Content,
					&insertedMessage.ReplyBody.Edited,
					&insertedMessage.ReplyBody.Reactions,
					&insertedMessage.ReplyBody.SentBy,
				); err != nil{
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					println(err.Error())
					return
				}
			}

			if body.Server == nil{
				websockets.HandleMessageSendWS(server, body.Target.String(), insertedMessage)
			}
			analytics.OnSentMessage()
			c.JSON(200, insertedMessage)
		}else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
			return
		}
	}
}
