package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"strings"
)

var AllowedChanges = map[string]bool{
	"desc":            true,
	"displayname":     true,
	"email":           true,
	"username":        true,
	"bannercolor":     true,
	"backgroundcolor": true,
}

func ChangeSelfInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define the structure of the expected request body
		var requestBody struct {
			Type     string `json:"type"`     // Field type to change (e.g., "email", "username")
			NewValue string `json:"newValue"` // New value for the specified field
		}

		// Extract JWT from the Authorization header
		jwt := c.GetHeader("Authorization")

		// Verify the JWT token
		if ret := security.VerifyJWT(jwt, session); ret.Status {

			// Parse the request body and handle JSON binding errors
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Request format is incorrect."})
				return
			}

			// Check if the requested change type is allowed
			if AllowedChanges[requestBody.Type] {

				// Special validation for email field
				if requestBody.Type == "email" {
					// Ensure the new email contains "@" and is of sufficient length
					if !strings.Contains(requestBody.NewValue, "@") || len(requestBody.NewValue) <= 3 {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Bad new value."})
						return
					}
				}

				// Define a structure to hold the user's information
				var requestUser struct {
					Createdat string `json:"createdat"` // Timestamp of account creation
					UserId    string `json:"user_id"`  // Unique user identifier
					Username  string `json:"username"` // Username of the user
				}

				// Fetch user details based on the provided JWT
				if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(
					&requestUser.Createdat,
					&requestUser.UserId,
					&requestUser.Username,
				); err != nil {
					// Return an error if the user cannot be found
					c.JSON(http.StatusBadRequest, gin.H{"error": "User not found."})
					return
				}

				// Update the specified field for the user
				if err := session.Query(
					`UPDATE users SET `+requestBody.Type+` = ? WHERE createdat = ? AND user_id = ? AND username = ?`,
					requestBody.NewValue,
					requestUser.Createdat,
					requestUser.UserId,
					requestUser.Username).Exec(); err != nil {
					// Handle errors during the database update
					println(err.Error())
					c.Status(http.StatusInternalServerError)
					return
				} else {
					// Return 200 on successful update
					c.Status(200)
					return
				}
			} else {
				// Return an error if the change type is not allowed
				c.JSON(http.StatusBadRequest, gin.H{"error": "Action not allowed by server."})
				return
			}
		} else {
			// Return 401 if the JWT token is invalid
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
			return
		}
	}
}

