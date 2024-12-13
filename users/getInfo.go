package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		target := c.Param("userId")
		var body struct {
			Profile *bool `json:"profile,omitempty"`
		}
		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId parameter is required"})
			return
		}
		err := c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		var user struct {
			Createdat       string       `json:"createdat"`
			UserId          gocql.UUID   `json:"user_id"`
			Username        string       `json:"username"`
			Badges          []gocql.UUID `json:"badges"`
			Displayname     string       `json:"displayname"`
			BannerColor     string       `json:"bannercolor"`
			BackgroundColor string       `json:"backgroundcolor"`
			Status          int          `json:"status"`
			Desc            string       `json:"desc"`
			Profile         *[]byte      `json:"profile"`
		}
		if body.Profile != nil && *body.Profile {
			if err := session.Query(
				`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc, profile FROM users WHERE user_id = ? ALLOW FILTERING`,
				target,
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
				&user.Profile,
			); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := session.Query(
				`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc FROM users WHERE user_id = ? ALLOW FILTERING`,
				target,
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
			); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, user)
		return
	}
}
