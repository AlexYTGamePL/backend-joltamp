package friends

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type friendsRemoveUser struct {
	friends   map[gocql.UUID]int8
	createdat string
	userId    gocql.UUID
	username  string
}

func RemoveFriend(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Checking JWT
		jwt := c.GetHeader("Authorization")
		if ret := security.VerifyJWT(jwt, session); ret.Status {
			var body struct {
				Target gocql.UUID `json:"target"`
			}
			if err := c.BindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var userSender friendsRemoveUser
			var userTarget friendsRemoveUser

			// Getting data about user
			if err := session.Query(`SELECT friends, createdat, user_id, username FROM users WHERE user_id = ? AND username = ? ALLOW FILTERING`, ret.User.UserId, ret.User.Username).Scan(
				&userSender.friends,
				&userSender.createdat,
				&userSender.userId,
				&userSender.username); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				println("remove_friends#1 " + err.Error())
				return
			}

			// Getting data about target
			if err := session.Query(`SELECT friends, createdat, user_id, username FROM users WHERE user_id = ? ALLOW FILTERING`, body.Target).Scan(
				&userTarget.friends,
				&userTarget.createdat,
				&userTarget.userId,
				&userTarget.username); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				println("remove_friends#2 " + err.Error())
				return
			}

			// Updating friends list for user
			if _, exists := userSender.friends[body.Target]; exists {
				delete(userSender.friends, body.Target)
				if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, userSender.friends, userSender.createdat, userSender.userId, userSender.username).Exec(); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					println("remove_friends#3 " + err.Error())
					return
				}
			} else {
				c.Status(http.StatusBadRequest)
				return
			}

			// Updating friends list for target
			if _, exists := userTarget.friends[ret.User.UserId]; exists {
				delete(userTarget.friends, ret.User.UserId)
				if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, userTarget.friends, userTarget.createdat, userTarget.userId, userTarget.username).Exec(); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					println("remove_friends#4 " + err.Error())
					return
				}
			} else {
				c.Status(http.StatusBadRequest)
				return
			}

			c.Status(200)

		} else {
			c.Status(http.StatusUnauthorized)
			return
		}
	}
}
