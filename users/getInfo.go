package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		target := c.Param("userId")
		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId parameter is required"})
			return
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
			Desc string `json:"desc"`
		}
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
		c.JSON(http.StatusOK, user)
		return
	}
}
