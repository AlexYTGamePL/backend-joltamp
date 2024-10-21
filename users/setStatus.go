package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func SetStatus(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt := c.GetHeader("Authorization")
		var body struct {
			Status int `json:"status"`
		}
		if ret := security.VerifyJWT(jwt, session); ret.Status {
			if err := c.BindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.Status == 0 || body.Status == 1 || body.Status == 2 || body.Status == 3 {
				var user struct {
					createdat string
					userId    gocql.UUID
					username  string
				}
				if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(&user.createdat, &user.userId, &user.username); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#002 - " + err.Error()})
					return
				}

				if err := session.Query(`UPDATE users SET status = ? WHERE createdat = ? AND user_id = ? AND username = ?`, body.Status, user.createdat, user.userId, user.username).Exec(); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#003 - " + err.Error()})
					return
				}

				c.Status(http.StatusOK)
				return

			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token!"})
			return
		}
	}
}
