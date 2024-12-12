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
		var requestBody struct {
			Type     string `json:"type"`
			NewValue string `json:"newValue"`
		}
		jwt := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(jwt, session); ret.Status {
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Request format is incorrect."})
				return
			}
			if AllowedChanges[requestBody.Type] {

				if requestBody.Type == "email" {
					if !strings.Contains(requestBody.NewValue, "@") || len(requestBody.NewValue) <= 3 {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Bad new value."})
					}
				}

				var requestUser struct {
					Createdat string `json:"createdat"`
					UserId    string `json:"user_id"`
					Username  string `json:"username"`
				}
				if err := session.Query(`SELECT createdat, user_id, username FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(
					&requestUser.Createdat,
					&requestUser.UserId,
					&requestUser.Username,
				); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "User not found."})
					return
				}
				if err := session.Query(
					`UPDATE users SET `+requestBody.Type+` = ? WHERE createdat = ? AND user_id = ? AND username = ?`,
					requestBody.NewValue,
					requestUser.Createdat,
					requestUser.UserId,
					requestUser.Username).Exec(); err != nil {
					println(err.Error())
					c.Status(http.StatusInternalServerError)
					return
				} else {
					c.Status(200)
					return
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Action not allowed by server."})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
			return
		}
	}
}
