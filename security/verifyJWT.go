package security

import "github.com/gocql/gocql"

func VerifyJWT(jwt string, session *gocql.Session) struct {
	Status bool
	User   struct {
		UserId      gocql.UUID
		Username    string
		Displayname string
	}
} {
	var ScanUser struct {
		UserId      gocql.UUID
		Username    string
		Displayname string
	}
	if err := session.Query(`SELECT user_id, username, displayname FROM users WHERE jwt = ? ALLOW FILTERING`, jwt).Scan(&ScanUser.UserId, &ScanUser.Username, &ScanUser.Displayname); err != nil {
		return struct {
			Status bool
			User   struct {
				UserId      gocql.UUID
				Username    string
				Displayname string
			}
		}{
			Status: false,
			User: struct {
				UserId      gocql.UUID
				Username    string
				Displayname string
			}{
				UserId:      gocql.UUID{},
				Username:    "",
				Displayname: "",
			},
		}
	} else {
		return struct {
			Status bool
			User   struct {
				UserId      gocql.UUID
				Username    string
				Displayname string
			}
		}{Status: true, User: ScanUser}
	}
}
