package analytics

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type DataLogs struct {
	Day                     string // Represents the specific day of the analytics log
	MessagesCount           int32  // Number of messages sent on that day
	RegisterCount           int32  // Number of new registrations on that day
	WebsocketConnectionCount int32 // Number of WebSocket connections on that day
}

func LoadAnalytics(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT from the Authorization header
		JWT := c.GetHeader("Authorization")

		// Verify the JWT token for authenticity
		if ret := security.VerifyJWT(JWT, session); !ret.Status {
			// Return 401 if the JWT is invalid
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
			return
		}

		// Extract the number of days parameter from the request
		daysCount := c.Param("days")

		// Query the database for analytics data with a limit based on the number of days
		iter := session.Query(`SELECT * FROM analytics LIMIT ?`, daysCount).Iter()

		// Prepare a slice to hold all analytics data
		var allData []DataLogs

		// Iterate through the query results
		for {
			var dayData DataLogs
			// Scan each row into the dayData struct
			if !iter.Scan(&dayData.Day, &dayData.MessagesCount, &dayData.RegisterCount, &dayData.WebsocketConnectionCount) {
				break // Exit the loop if there are no more rows
			}
			// Append the data for the current day to the results slice
			allData = append(allData, dayData)
		}

		// Return the collected analytics data as a JSON response
		c.JSON(200, allData)
	}
}
