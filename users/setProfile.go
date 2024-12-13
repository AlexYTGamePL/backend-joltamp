package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"io/ioutil"
	"net/http"
)

func SetProfile(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the JWT token from the Authorization header
		JWT := c.GetHeader("Authorization")

		// Verify the JWT token
		if ret := security.VerifyJWT(JWT, session); !ret.Status {
			// Respond with 401 if the JWT token is invalid
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token"})
			return
		}

		// Struct to hold user details fetched from the database
		var requestUser struct {
			Createdat string `json:"createdat"`
			UserId    string `json:"user_id"`
			Username  string `json:"username"`
		}

		// Fetch the user details based on the JWT token
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, JWT).Scan(
			&requestUser.Createdat,
			&requestUser.UserId,
			&requestUser.Username,
		); err != nil {
			// Respond with 400 if the user is not found
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found."})
			return
		}

		// Retrieve the uploaded file from the request
		file, err := c.FormFile("photo")
		if err != nil {
			// Respond with 400 if the file cannot be loaded
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo can't be loaded."})
			return
		}

		// Check if the file size exceeds the maximum limit of 2MB
		if file.Size > 2097152 {
			// Respond with 400 for file size violation
			c.JSON(http.StatusBadRequest, gin.H{"error": "Max file size is 2MB"})
			return
		}

		// Open the uploaded file to read its contents
		fileContent, err := file.Open()
		if err != nil {
			// Respond with 400 if the file cannot be opened
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open file."})
			return
		}
		defer fileContent.Close() // Ensure the file is closed after processing

		// Read the content of the uploaded file
		photoData, err := ioutil.ReadAll(fileContent)
		if err != nil {
			// Respond with 400 if file content cannot be read
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read file content."})
			return
		}

		// Update the user's profile photo in the database
		if err := session.Query(`UPDATE users SET profile = ? WHERE createdat = ? AND user_id = ? AND username = ?`,
			photoData,
			requestUser.Createdat,
			requestUser.UserId,
			requestUser.Username,
		).Exec(); err != nil {
			// Respond with 500 if there's an error updating the database
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile photo."})
			return
		}

		// Respond with 200 OK on successful update
		c.Status(http.StatusOK)
		return
	}
}
