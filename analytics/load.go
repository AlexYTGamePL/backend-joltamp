package analytics

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type DataLogs struct {
	Day string
	MessagesCount int32
	RegisterCount int32
	WebsocketConnectionCount int32
}

func LoadAnalytics(session *gocql.Session) gin.HandlerFunc{
	return func(c *gin.Context) {
		JWT := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(JWT, session); !ret.Status {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
			return
		}
		daysCount := c.Param("days")
		iter := session.Query(`SELECT * FROM analytics LIMIT ?`, daysCount).Iter()
		var allData []DataLogs
		for {
			var dayData DataLogs
			if !iter.Scan(&dayData.Day, &dayData.MessagesCount, &dayData.RegisterCount, &dayData.WebsocketConnectionCount) {
				break
			}
			allData = append(allData, dayData)
		}
		c.JSON(200, allData)
	}
}