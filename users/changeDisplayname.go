package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type requestType struct {
	jwt            gocql.UUID `json:"jwt"`
	newDisplayname string     `json:"displayname"`
}
type userType struct {
	createdat string
	userId    gocql.UUID
	username  string
}

func ChangeDisplayname(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		request := requestType{}
		if err := c.BindJSON(request); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Wrong request format"})
			return
		}

		dbuser := userType{}
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ?`, request.jwt).Scan(&dbuser.createdat, &dbuser.userId, &dbuser.username); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		if err := session.Query(`UPDATE users SET displayname = ? WHERE createdat = ? AND user_id = ? AND username = ?`, request.newDisplayname, dbuser.createdat, dbuser.userId, dbuser.username).Exec(); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		c.Status(http.StatusOK)
		return
	}
}
