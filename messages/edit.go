package messages

import (
    "backend-joltamp/security"
    "backend-joltamp/types"
    "backend-joltamp/websockets"
    "github.com/gin-gonic/gin"
    "github.com/gocql/gocql"
    "net/http"
)

func EditMessage(session *gocql.Session) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract the JWT token from the Authorization header
        jwt := c.GetHeader("Authorization")
        // Verify the JWT token and ensure it is valid
        if ret := security.VerifyJWT(jwt, session); ret.Status {

            // Define the request body structure
            var body struct {
                Target     gocql.UUID  `json:"target"`
                Server     *gocql.UUID `json:"server"`
                Message    gocql.UUID  `json:"message"`
                SentAt     string      `json:"sentat"`
                SentAtTime int64       `json:"sentattime"`
                Content string `json:"content"`
            }
            // Bind JSON body to the struct and handle parsing errors
            if err := c.BindJSON(&body); err != nil {
                c.Status(http.StatusBadRequest)
                return
            }

            // Determine the target, with special handling if the server ID is nil
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

                // Check if the message was sent by the authenticated user
                var sentby gocql.UUID
                if err := session.Query(`SELECT sent_by FROM messages WHERE server_id = ? AND message_id = ? AND sent_at = ? AND target_id = ? AND sent_at_time = ?`,
                    server,
                    body.Message,
                    body.SentAt,
                    target,
                    body.SentAtTime,
                ).Scan(&sentby); err != nil {
                    // Return 500 if the message is not found or a database error occurs
                    println(err.Error())
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Message not found or Internal server Error"})
                    return
                }
                // If the authenticated user is not the sender, return 401
                if sentby != ret.User.UserId {
                    c.JSON(http.StatusUnauthorized, gin.H{"error": "You cant edit message sent by someone else"})
                    return
                }

            } else {
                target = body.Target.String()
            }

            // Update the message content and mark it as edited in the database
            if err := session.Query(`UPDATE messages SET content = ?, edited = ? WHERE server_id = ? AND message_id = ? AND sent_at = ? AND target_id = ? AND sent_at_time = ?`,
                body.Content,
                true,
                server,
                body.Message,
                body.SentAt,
                target,
                body.SentAtTime,
            ).Exec(); err != nil {
                c.Status(http.StatusInternalServerError)
                return
            }

            // Notify connected clients about the message edit via WebSocket
            websockets.HandleMessageEditWS(server, body.Target.String(), types.EditMessage{
                ServerId:   server,
                TargetId:   target,
                SentAt:     body.SentAt,
                SentAtTime: body.SentAtTime,
                MessageId:  body.Message,
                Content:    body.Content,
            });
            c.Status(http.StatusOK)
            return

        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
            return
        }
    }
}