package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func ChangeDisplayname(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			NewDisplayname string `json:"displayname"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Wrong request format"})
			println(err.Error())
			return
		}
		jwt := c.GetHeader("Authorization")
		var dbuser struct {
			createdat string
			userId    gocql.UUID
			username  string
		}
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(&dbuser.createdat, &dbuser.userId, &dbuser.username); err != nil {
			println(err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		if err := session.Query(`UPDATE users SET displayname = ? WHERE createdat = ? AND user_id = ? AND username = ?`, request.NewDisplayname, dbuser.createdat, dbuser.userId, dbuser.username).Exec(); err != nil {
			println(err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		c.Status(http.StatusOK)
		return
	}
}
