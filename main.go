package main

import (
	"backend-joltamp/friends"
	"backend-joltamp/messages"
	"backend-joltamp/users"
	"log"
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
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	apiV0 := router.Group("/api/v0")
	// User login/register

	apiV0.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "All services working as usual!"})
		return
	})

	usersRouter := apiV0.Group("/users")
	usersRouter.POST("/register", users.SaveUser(session))
	usersRouter.POST("/login", users.GetUser(session))
	usersRouter.POST("/changeDisplayname", users.ChangeDisplayname(session))
	usersRouter.POST("/changeEmail", users.ChangeEmail(session))
	usersRouter.GET("/isAdmin/:userId", users.IsAdmin(session))
	usersRouter.GET("/getInfo/:userId", users.GetInfo(session))
	usersRouter.POST("/setStatus", users.SetStatus(session))
	// User friends
	friendsRouter := apiV0.Group("/friends")
	friendsRouter.POST("/", friends.GetFriends(session))
	friendsRouter.POST("/remove", friends.RemoveFriend(session))
	friendsRouter.POST("/sendRequest", friends.SendRequest(session))

	messagesRouter := apiV0.Group("/messages")
	messagesRouter.POST("/", messages.LoadMessages(session))
	messagesRouter.POST("/send", messages.SendMessage(session))
	messagesRouter.POST("/delete", messages.DeleteMessage(session))

	// Starting REST API on HTTPS or HTTP depending on .env variable
	if os.Getenv("RUN_AS_HTTPS") == "true" {

		// Running on HTTPS
		println("Running as HTTPS")
		err := router.RunTLS(os.Getenv("BACKEND_RUN_IP")+":"+os.Getenv("BACKEND_RUN_PORT"), os.Getenv("BACKEND_CERT"), os.Getenv("BACKEND_KEY"))

		// Error handling
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
	} else {

		// Running on HTTP
		println("Running as HTTP")
		err := router.Run(os.Getenv("BACKEND_RUN_IP") + ":" + os.Getenv("BACKEND_RUN_PORT"))

		// Error handling
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
	}
}
