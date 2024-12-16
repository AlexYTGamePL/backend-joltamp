package friends

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

type user struct {
	FriendStatus    *int8        `json:"friendstatus,omitempty"`
	Createdat       string       `json:"createdat"`
	UserId          gocql.UUID   `json:"user_id"`
	Username        string       `json:"username"`
	Badges          []gocql.UUID `json:"badges"`
	Displayname     string       `json:"displayname"`
	BannerColor     string       `json:"bannercolor"`
	BackgroundColor string       `json:"backgroundcolor"`
	Status          int8         `json:"status"`
	Desc string `json:"desc"`
}

func GetFriends(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		var friends map[gocql.UUID]int8

		jwt := c.GetHeader("Authorization")

		if err := session.Query(`SELECT friends FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Consistency(gocql.One).Scan(&friends); err != nil {
			println(friends)
			println(err.Error())
			return
		} else {
			result := make(map[gocql.UUID]user)
			for uuid, status := range friends {
				var userDetail user
				if err := session.Query(
					`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc FROM users WHERE user_id = ? ALLOW FILTERING`,
					uuid,
				).Scan(
					&userDetail.Createdat,
					&userDetail.UserId,
					&userDetail.Username,
					&userDetail.Badges,
					&userDetail.Displayname,
					&userDetail.BannerColor,
					&userDetail.BackgroundColor,
					&userDetail.Status,
					&userDetail.Desc,
				); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				} else {
					userDetail.FriendStatus = &status
					result[userDetail.UserId] = userDetail
				}
				continue
			}
			c.JSON(http.StatusOK, result)
		}
	}
}
