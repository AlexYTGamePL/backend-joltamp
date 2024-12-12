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
		JWT := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(JWT, session); !ret.Status{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token"})
			return
		}
		var requestUser struct {
			Createdat string `json:"createdat"`
			UserId    string `json:"user_id"`
			Username  string `json:"username"`
		}
		if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, JWT).Scan(
			&requestUser.Createdat,
			&requestUser.UserId,
			&requestUser.Username,
		); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found."})
			return
		}
		file, err := c.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo cant be loaded."})
			return
		}
		if file.Size > 2097152 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Max file size is 2MB"})
		}
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot open file."})
			return
		}
		defer fileContent.Close()

		photoData, err := ioutil.ReadAll(fileContent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read file content."})
			return
		}
		println(file.Size)
		if err := session.Query(`UPDATE users SET profile = ? WHERE createdat = ? AND user_id = ? AND username = ?`,
			photoData,
			requestUser.Createdat,
			requestUser.UserId,
			requestUser.Username,
		).Exec(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		c.Status(200)
		return
	}
}