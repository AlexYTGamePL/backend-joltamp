package friends

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type friendsUser struct {
	createdat string
	userId    gocql.UUID
	username  string
	friends   map[gocql.UUID]int8
}

func SendRequest(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			To     string `json:"to"`
			Action *bool  `json:"action"`
		}
		var sender friendsUser
		var target friendsUser
		jwt := c.GetHeader("Authorization")
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "friendsRequest#001 - " + err.Error()})
			return
		}
		if err := session.Query(`SELECT createdat, user_id, username, friends FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(&sender.createdat, &sender.userId, &sender.username, &sender.friends); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#002 - " + err.Error()})
			return
		}
		if err := session.Query(`SELECT createdat, user_id, username, friends FROM users WHERE username = ? ALLOW FILTERING`, request.To).Scan(&target.createdat, &target.userId, &target.username, &target.friends); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#003 - " + err.Error()})
			return
		}
		if ele, exists := sender.friends[target.userId]; exists {
			if ele == 1 && target.friends[sender.userId] == 0 {
				if request.Action != nil {
					if *request.Action {
						sender.friends[target.userId] = 2
						if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, sender.friends, sender.createdat, sender.userId, sender.username).Exec(); err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#004 - " + err.Error()})
							return
						}
						target.friends[sender.userId] = 2
						if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, target.friends, target.createdat, target.userId, target.username).Exec(); err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#005 - " + err.Error()})
							return
						}
					} else {
						delete(sender.friends, target.userId)
						if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, sender.friends, sender.createdat, sender.userId, sender.username).Exec(); err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#004 - " + err.Error()})
							return
						}
						delete(target.friends, sender.userId)
						if err := session.Query(`UPDATE users SET friends = ? WHERE createdat = ? AND user_id = ? AND username = ?`, target.friends, target.createdat, target.userId, target.username).Exec(); err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#005 - " + err.Error()})
							return
						}
					}
				} else {
					c.Status(http.StatusBadRequest)
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Wrong fields set!"})
				return
			}
		} else {
			if err := session.Query(`UPDATE users SET friends = friends + {?: ?} WHERE createdat = ? AND user_id = ? AND username = ?`, target.userId, 0, sender.createdat, sender.userId, sender.username).Exec(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#006 - " + err.Error()})
				return
			}
			if err := session.Query(`UPDATE users SET friends = friends + {?: ?} WHERE createdat = ? AND user_id = ? AND username = ?`, sender.userId, 1, target.createdat, target.userId, target.username).Exec(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "friendsRequest#007 - " + err.Error()})
				return
			}
		}
		c.Status(http.StatusOK)
		return
	}
}
