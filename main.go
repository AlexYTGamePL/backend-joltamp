package main

import (
	"backend-joltamp/analytics"
	"backend-joltamp/friends"
	"backend-joltamp/messages"
	"backend-joltamp/users"
	"backend-joltamp/websockets"
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))
	apiV0 := router.Group("/api/v0")

	apiV0.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "All services working as usual!"})
		return
	})

	// Users routes
	usersRouter := apiV0.Group("/users")
	usersRouter.POST("/register", users.SaveUser(session))
	usersRouter.POST("/login", users.GetUser(session))
	usersRouter.POST("/changeDisplayname", users.ChangeDisplayname(session))
	usersRouter.POST("/changeEmail", users.ChangeEmail(session))
	usersRouter.GET("/isAdmin/:userId", users.IsAdmin(session))
	usersRouter.GET("/getInfo/:userId", users.GetInfo(session))
	usersRouter.POST("/getSelfInfo", users.GetSelfInfo(session))
	usersRouter.POST("/setStatus", users.SetStatus(session))
	usersRouter.POST("/changeSelfInfo", users.ChangeSelfInfo(session))
	usersRouter.POST("/setProfile", users.SetProfile(session))
	// Friends routes
	friendsRouter := apiV0.Group("/friends")
	friendsRouter.POST("/", friends.GetFriends(session))
	friendsRouter.POST("/remove", friends.RemoveFriend(session))
	friendsRouter.POST("/sendRequest", friends.SendRequest(session))

	messagesRouter := apiV0.Group("/messages")
	messagesRouter.POST("/", messages.LoadMessages(session))
	messagesRouter.POST("/send", messages.SendMessage(session))
	messagesRouter.POST("/delete", messages.DeleteMessage(session))
	messagesRouter.POST("/edit", messages.EditMessage(session))

	webSocket := apiV0.Group("/ws")
	webSocket.GET("/", websockets.WebsocketHandler(session))

	analyticsRouter := apiV0.Group("/analytics")
	analyticsRouter.GET("/:days", analytics.LoadAnalytics(session))

	// Starting REST API on HTTPS or HTTP depending on .env variable
	analytics.Main(session);
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
