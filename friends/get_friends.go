package friends

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

func GetFriends(sesion *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		var friends map[gocql.UUID]bool

		jwt := c.GetHeader("Authorization")

		if err := sesion.Query(`SELECT friends_ids FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Consistency(gocql.One).Scan(&friends); err != nil {
			println(friends)
			println(err.Error())
			return
		} else {
			c.IndentedJSON(http.StatusOK, gin.H{"friends": friends})
			return
		}
	}
}
