package analytics

import (
	"github.com/gocql/gocql"
	"log"
	"time"
)

var SentMessagesCounter int32
var RegisteredAccountsCounter int32
var WebsocketsConnectionsCounter int32

func Main(session *gocql.Session) {
	go func() {

		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day() + 1, 0, 0, 0, 0, now.Location())
			duration := time.Until(next)
			time.Sleep(duration)
			println("Saving analytics data...")
			SaveData(session)
			SentMessagesCounter = 0
			RegisteredAccountsCounter = 0
			WebsocketsConnectionsCounter = 0
			println("Analytics data has been saved.")
		}
	}()
}

func SaveData(session *gocql.Session) {
	now := time.Now()
	if err := session.Query(`INSERT INTO analytics (day, messagescount, registercount, wsconnectcount) VALUES (?, ?, ?, ?)`,
		now.Format("2006-01-02"), SentMessagesCounter, RegisteredAccountsCounter, WebsocketsConnectionsCounter).Exec(); err != nil{
		log.Fatalln("Error while instrting analytics!")
		return
	}
}

func OnSentMessage() {
	SentMessagesCounter++
}
func OnRegisterAccount() {
	RegisteredAccountsCounter++
}
func OnWebsocketsConnect() {
	WebsocketsConnectionsCounter++
}