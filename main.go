package main

import (
	"backend-joltamp/friends"
	"backend-joltamp/users"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
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
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// User login/register
	router.POST("/users/register", users.SaveUser(session))
	router.POST("/users/login", users.GetUser(session))
	// User friends
	router.GET("/friends/:jwt", friends.GetFriends(session))
	errtwo := router.Run("192.168.0.20:3000")
	if errtwo != nil {
		println(errtwo.Error())
		return
	}
}
