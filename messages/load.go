package messages

import (
	"backend-joltamp/security"
	"backend-joltamp/types"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func LoadMessages(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(jwt, session); ret.Status {
			var body struct {
				Target gocql.UUID  `json:"target"`
				Server *gocql.UUID `json:"server"`
				Latest *int64     `json:"latest"`
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
			var server string
			if body.Server == nil {
				server = ""
			} else {
				server = body.Server.String()
			}
			var iter *gocql.Iter
			if body.Latest != nil {
				iter = session.Query(`SELECT * FROM messages WHERE target_id = ? AND server_id = ? AND sent_at_time < ? LIMIT 50 ALLOW FILTERING`, target, server, body.Latest).Iter()
			} else {
				iter = session.Query(`SELECT * FROM messages WHERE target_id = ? AND server_id = ? LIMIT 50`, target, server).Iter()
			}
			var messages []types.Message
			println(iter.NumRows())
			for {
				var msg types.Message

				// Scan the current row into msg
				if !iter.Scan(&msg.ServerId, &msg.TargetId, &msg.SentAt, &msg.SentAtTime, &msg.MessageId, &msg.Content, &msg.Edited, &msg.Reactions, &msg.Reply, &msg.SentBy) {
					break
				}
				if msg.Reply != "" {
					if msg.ReplyBody == nil {
						msg.ReplyBody = &types.ReplyBodyType{}
					}
					if err := session.Query(`SELECT server_id, target_id, sent_at, sent_at_time, message_id, content, edited, reactions, sent_by FROM messages WHERE target_id = ? AND server_id = ? AND message_id = ? ALLOW FILTERING`, target, server, msg.Reply).Scan(
						&msg.ReplyBody.ServerId,
						&msg.ReplyBody.TargetId,
						&msg.ReplyBody.SentAt,
						&msg.ReplyBody.SentAtTime,
						&msg.ReplyBody.MessageId,
						&msg.ReplyBody.Content,
						&msg.ReplyBody.Edited,
						&msg.ReplyBody.Reactions,
						&msg.ReplyBody.SentBy); err != nil {
						println(err.Error())
						c.JSON(http.StatusInternalServerError, err.Error())
					}
				}
				messages = append(messages, msg)
			}
			reversedMessages := make([]types.Message, len(messages))
			if len(messages) > 1{
				for i, j := 0, len(messages)-1; i <= j; i, j = i+1, j-1 {
					reversedMessages[i] = messages[j]
					reversedMessages[j] = messages[i]
				}
			}else{
				reversedMessages = messages
			}
			c.JSON(http.StatusOK, reversedMessages)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
			return
		}
	}
}
