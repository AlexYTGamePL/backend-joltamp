package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type changeEmailRequestType struct {
	email string `json:"email"`
}

func ChangeEmail(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request changeEmailRequestType
		var dbuser userType
		jwt := c.GetHeader("Authorization")

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong request format"})
			return
		}

		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Consistency(gocql.One).Scan(&dbuser.createdat, &dbuser.userId, &dbuser.username); err != nil {
			println(err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		if err := session.Query(`UPDATE users SET email = ? WHERE createdat = ? AND user_id = ? AND username = ?`, request.email, dbuser.createdat, dbuser.userId, dbuser.username).Exec(); err != nil {
			println(err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		c.Status(http.StatusOK)
		return
	}
}
