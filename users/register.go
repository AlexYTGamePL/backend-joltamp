package users

import (
	"backend-joltamp/analytics"
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"strings"
)

func SaveUser(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newUser struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
		}
		if err := c.BindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect body format"})
			return
		}
		newUser.Email = strings.ToLower(newUser.Email)
		if len(newUser.Username) <= 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is incorrect (>4)"})
			return
		}
		if len(newUser.Password) <= 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is incorrect (>6)"})
			return
		}
		if len(newUser.Email) <= 3 && strings.Contains(newUser.Email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is incorrect (>3)"})
			return
		}
		var id gocql.UUID
		if err := session.Query(`SELECT user_id FROM users WHERE username = ? ALLOW FILTERING`, newUser.Username).Scan(&id); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this username already exists"})
			return
		}
		if err := session.Query(`SELECT user_id FROM users WHERE email = ? ALLOW FILTERING`, newUser.Email).Scan(&id); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
			return
		}
		passwordHash, hashErr := security.HashPassword(newUser.Password)
		if hashErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error p#001"})
			return
		}
		userJwt, _ := gocql.RandomUUID()
		userId, _ := gocql.RandomUUID()
		err := session.Query(`INSERT INTO users (createdat, user_id, username, displayname, email, password, isadmin, jwt, status) VALUES (todate(now()), ?, ?, ?, ?, ?, false, ?, 0)`, userId, newUser.Username, newUser.Username, newUser.Email, passwordHash, userJwt).Exec()
		if err != nil {
			println(err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server error db#001"})
			return
		}
		analytics.OnRegisterAccount()
		c.JSON(http.StatusOK, gin.H{
			"message": "Created",
			"user": struct {
				Username string
				Email    string
				JWT      gocql.UUID
				UserId   gocql.UUID
			}{
				Username: newUser.Username,
				Email:    newUser.Email,
				JWT:      userJwt,
				UserId:   userId,
			},
		})
	}
}
