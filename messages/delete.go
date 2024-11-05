package messages

import (
    "backend-joltamp/security"
    "github.com/gin-gonic/gin"
    "github.com/gocql/gocql"
    "net/http"
)

func DeleteMessage(session *gocql.Session) gin.HandlerFunc {
    return func(c *gin.Context) {
        jwt := c.GetHeader("Authorization")
        if ret := security.VerifyJWT(jwt, session); ret.Status {
            var body struct {
                Target     gocql.UUID  `json:"target"`
                Server     *gocql.UUID `json:"server"`
                Message    gocql.UUID  `json:"message"`
                SentAt     string      `json:"sentat"`
                SentAtTime int64       `json:"sentattime"`
            }

            if err := c.BindJSON(&body); err != nil {
                c.Status(http.StatusBadRequest)
                return
            }
            var target string
            server := func() string {
                if body.Server == nil {
                    return ""
                } else {
                    return body.Server.String()
                }
            }()
            if server == "" {
                target = CombineUUIDs(ret.User.UserId, body.Target)
                var sentby gocql.UUID
                if err := session.Query(`SELECT sent_by FROM messages WHERE server_id = ? AND message_id = ? AND sent_at = ? AND target_id = ? AND sent_at_time = ?`,
                    server,
                    body.Message,
                    body.SentAt,
                    target,
                    body.SentAtTime,
                ).Scan(&sentby); err != nil {
                    println(err.Error())
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Message not found or Internal server Error"})
                    return
                }

                if sentby != ret.User.UserId {
                    c.JSON(http.StatusUnauthorized, gin.H{"error": "You cant delete message sent by someone else"})
                    return
                }

            } else {
                target = body.Target.String()
            }
            if err := session.Query(`DELETE FROM messages WHERE server_id = ? AND message_id = ? AND sent_at = ? AND target_id = ? AND sent_at_time = ?`,
                server,
                body.Message,
                body.SentAt,
                target,
                body.SentAtTime,
            ).Exec(); err != nil {
                c.Status(http.StatusInternalServerError)
                return
            }

            c.Status(http.StatusOK)
            return

        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
            return
        }
    }
}
