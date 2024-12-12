package users

import (
	"backend-joltamp/security"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
)

func GetSelfInfo(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		JWT := c.GetHeader("Authorization")
			if ret := security.VerifyJWT(JWT, session); !ret.Status {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect JWT token."})
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
			Email string `json:"email"`
		}
		if err := session.Query(
			`SELECT createdat, user_id, username, badges, displayname, bannercolor, backgroundcolor, status, desc, email FROM users WHERE jwt = ? ALLOW FILTERING`,
			JWT,
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
			&user.Email,
		); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
		return
	}
}
