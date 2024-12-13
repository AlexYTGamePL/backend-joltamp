package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func IsAdmin(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("userId")
		var isAdmin bool
		if err := session.Query(`SELECT isAdmin FROM users WHERE user_id = ? ALLOW FILTERING`, userId).Scan(&isAdmin); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			c.IndentedJSON(http.StatusOK, isAdmin)
			return
		}
	}
}
