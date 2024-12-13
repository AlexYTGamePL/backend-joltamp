package analytics

import (
	"github.com/gocql/gocql"
	"log"
	"time"
)

// Global counters for tracking analytics data
var SentMessagesCounter int32               // Tracks the number of messages sent
var RegisteredAccountsCounter int32         // Tracks the number of accounts registered
var WebsocketsConnectionsCounter int32      // Tracks the number of WebSocket connections

// Main initializes a goroutine to save analytics data daily
func Main(session *gocql.Session) {
	go func() {
		for {
			// Calculate the duration until midnight (start of the next day)
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := time.Until(next)

			// Sleep until the calculated duration elapses
			time.Sleep(duration)

			// Log the start of saving analytics data
			println("Saving analytics data...")
			SaveData(session) // Save the current analytics data

			// Reset the counters after data is saved
			SentMessagesCounter = 0
			RegisteredAccountsCounter = 0
			WebsocketsConnectionsCounter = 0
			println("Analytics data has been saved.") // Log completion
		}
	}()
}

// SaveData inserts the current analytics data into the database
func SaveData(session *gocql.Session) {
	// Get the current date in "YYYY-MM-DD" format
	now := time.Now()

	// Execute the CQL query to insert the analytics data
	if err := session.Query(`INSERT INTO analytics (day, messagescount, registercount, wsconnectcount) VALUES (?, ?, ?, ?)`,
		now.Format("2006-01-02"), SentMessagesCounter, RegisteredAccountsCounter, WebsocketsConnectionsCounter).Exec(); err != nil {
		// Log a fatal error if the query fails
		log.Fatalln("Error while inserting analytics!")
		return
	}
}

// OnSentMessage increments the message counter when a message is sent
func OnSentMessage() {
	SentMessagesCounter++
}

// OnRegisterAccount increments the account registration counter when a new account is registered
func OnRegisterAccount() {
	RegisteredAccountsCounter++
}

// OnWebsocketsConnect increments the WebSocket connection counter when a new WebSocket connection is established
func OnWebsocketsConnect() {
	WebsocketsConnectionsCounter++
}
