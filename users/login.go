package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetUser(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := c.BindJSON(&user)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Incorrect body format"})
			return
		}
		if len(user.Email) <= 3 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Email is incorrect"})
			return
		}
		if len(user.Password) <= 6 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Password is incorrect"})
			return
		}
		var dbUser struct {
			UserID   gocql.UUID
			Password string
		}
		if err := session.Query(`SELECT user_id, password FROM users WHERE email = ? ALLOW FILTERING`, user.Email).Scan(&dbUser.UserID, &dbUser.Password); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "User not found"})
			return
		}

		if security.CheckPasswordHash(user.Password, dbUser.Password) {
			c.IndentedJSON(http.StatusOK, gin.H{"user": dbUser.UserID})
			return
		} else {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
			return
		}
	}
}
