package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the 'userId' from the URL parameters
		target := c.Param("userId")

		// Define a struct to hold the optional profile query flag
		var body struct {
			Profile *bool `json:"profile,omitempty"`
		}

		// Check if the 'userId' parameter is provided
		if target == "" {
			// Return an error if 'userId' is not provided in the URL
			c.JSON(http.StatusBadRequest, gin.H{"error": "GetInfo#001 " +"userId parameter is required"})
			return
		}

		// Bind the incoming JSON body to the 'body' struct
		err := c.BindJSON(&body)
		if err != nil {
			// Return an error if there’s an issue with the JSON format
			c.JSON(http.StatusBadRequest, gin.H{"error": "GetInfo#002 " + err.Error()})
			return
		}

		// Define the struct that will hold the user details from the database
		var user struct {
			Createdat       string       `json:"createdat"`
			UserId          gocql.UUID   `json:"user_id"`
			Username        string       `json:"username"`
			Badges          []gocql.UUID `json:"badges"`
			Displayname     string       `json:"displayname"`
			BannerColor     string       `json:"bannercolor"`
			BackgroundColor string       `json:"backgroundcolor"`
			Status          int          `json:"status"`
			Desc            string       `json:"desc"`
			Profile         *[]byte      `json:"profile"`
		}

		// Check if the Profile flag is set and true in the request body
		if body.Profile != nil && *body.Profile {
			// Query the database to retrieve user details along with the profile image
			if err := session.Query(
				`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc, profile
				FROM users WHERE user_id = ? ALLOW FILTERING`,
				target,
			).Scan(
				&user.Createdat,
				&user.UserId,
				&user.Username,
				&user.Badges,
				&user.Displayname,
				&user.BannerColor,
				&user.BackgroundColor,
				&user.Status,
				&user.Desc,
				&user.Profile,
			); err != nil {
				// Return an error if there is an issue with the database query
				c.JSON(http.StatusBadRequest, gin.H{"error": "GetInfo#003 " +err.Error()})
				return
			}
		} else {
			// Query the database to retrieve user details without the profile image
			if err := session.Query(
				`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc
				FROM users WHERE user_id = ? ALLOW FILTERING`,
				target,
			).Scan(
				&user.Createdat,
				&user.UserId,
				&user.Username,
				&user.Badges,
				&user.Displayname,
				&user.BannerColor,
				&user.BackgroundColor,
				&user.Status,
				&user.Desc,
			); err != nil {
				// Return an error if there is an issue with the database query
				c.JSON(http.StatusBadRequest, gin.H{"error":"GetInfo#004 " + err.Error()})
				return
			}
		}

		// Set the response Content-Type to JSON
		c.Header("Content-Type", "application/json; charset=utf-8")

		// Return the user data as JSON in the response
		c.JSON(http.StatusOK, user)
		return
	}
}

