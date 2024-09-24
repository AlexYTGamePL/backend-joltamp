package main

import (
	"backend-joltamp/friends"
	"backend-joltamp/users"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		println(err.Error())
		return
	}

	scylladb := gocql.NewCluster(os.Getenv("SCYLLA_SERVER_IP") + ":" + os.Getenv("SCYLLA_SERVER_PORT"))
	scylladb.Keyspace = "joltamp"
	scylladb.Consistency = gocql.One

	session, err := scylladb.CreateSession()
	if err != nil {
		println(err.Error())
		return
	}
	defer session.Close()
	router := gin.Default()
	// User login/register
	router.POST("/users/register", users.SaveUser(session))
	router.POST("/users/login", users.GetUser(session))
	// User friends
	router.GET("/friends/:jwt", friends.GetFriends(session))
	errtwo := router.Run("localhost:3000")
	if errtwo != nil {
		println(errtwo.Error())
		return
	}
}