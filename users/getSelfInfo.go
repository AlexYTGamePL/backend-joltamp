package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetSelfInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT from the Authorization header
		JWT := c.GetHeader("Authorization")

		// Verify the JWT token for authenticity
		if ret := security.VerifyJWT(JWT, session); !ret.Status {
			// Return 401 if the JWT is invalid
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
			return
		}

		// Define a structure to hold the user information
		var user struct {
			Createdat       string       `json:"createdat"`       // Timestamp of account creation
			UserId          gocql.UUID   `json:"user_id"`         // Unique user identifier
			Username        string       `json:"username"`        // Username
			Badges          []gocql.UUID `json:"badges"`          // List of badge IDs
			Displayname     string       `json:"displayname"`     // User's display name
			BannerColor     string       `json:"bannercolor"`     // Color of the banner
			BackgroundColor string       `json:"backgroundcolor"` // Background color of the profile
			Status          int          `json:"status"`          // Status code (e.g., active, inactive)
			Desc            string       `json:"desc"`            // User's description or bio
			Email           string       `json:"email"`           // User's email address
			Profile         []byte       `json:"profile"`         // Profile picture data in binary format
		}

		// Query the database to fetch the user's details based on the provided JWT
		if err := session.Query(
			`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc, email, profile FROM users WHERE jwt = ? ALLOW FILTERING`,
			JWT,
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
			&user.Email,
			&user.Profile,
		); err != nil {
			// Handle errors during database query
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Return the user's information as a JSON response
		c.JSON(http.StatusOK, user)
		return
	}
}

