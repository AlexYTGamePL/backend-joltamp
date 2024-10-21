package friends

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func SendRequest(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			To string `json:"to"`
		}
		var sender struct {
			createdat string
			userId    gocql.UUID
			username  string
		}
		var target struct {
			createdat string
			userId    gocql.UUID
			username  string
		}
		jwt := c.GetHeader("Authorization")
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "friendsRequest#001 - " + err.Error()})
			return
		}
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(&sender.createdat, &sender.userId, &sender.username); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#002 - " + err.Error()})
			return
		}
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE user_id = ? ALLOW FILTERING`, request.To).Scan(&target.createdat, &target.userId, &target.username); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#003 - " + err.Error()})
			return
		}
		if err := session.Query(`UPDATE users SET friends = friends + {?: ?} WHERE createdat = ? AND user_id = ? AND username = ?`, request.To, 0, sender.createdat, sender.userId, sender.username).Exec(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#004 - " + err.Error()})
			return
		}
		if err := session.Query(`UPDATE users SET friends = friends + {?: ?} WHERE createdat = ? AND user_id = ? AND username = ?`, sender.userId, 1, target.createdat, target.userId, target.username).Exec(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#005 - " + err.Error()})
			return
		}

		c.Status(http.StatusOK)
		return
	}
}
