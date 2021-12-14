package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	shell "github.com/ipfs/go-ipfs-api"

	"git.kiefte.eu/lapingvino/infodump/message"
)

var DatabasePath = "infodump.db"
var DB *sql.DB

// StartOLNListener starts a PubSub listener that listens for messages from the network
// and adds them to LocalMessages
// For this it uses the IPFS gateway and listens on the topic "OLN", as well as
// the list of followed tags from the database
func StartOLNListener() {
	// Get the IPFS gateway
	gateway := message.IPFSGateway
	// Get the database
	db := GetDatabase()
	// Get the list of followed tags
	followedTags := GetFollowedTags(db)
	// Create a new IPFS client
	myIPFS := shell.NewShell(gateway)
	var subs []*shell.PubSubSubscription
	olnsub, err := myIPFS.PubSubSubscribe("OLN")
	if err != nil {
		fmt.Println(err)
		return
	}
	subs = append(subs, olnsub)
	for _, tag := range followedTags {
		tagssub, err := myIPFS.PubSubSubscribe("oln-" + tag)
		if err != nil {
			fmt.Println(err)
		} else {
			subs = append(subs, tagssub)
		}
	}
	// Start a goroutine for each of the subscriptions in subs,
	// read the CID from the Next method, look up the CID on IPFS,
	// read this in via message.MessagesFromIPFS and add the message to LocalMessages
	for _, sub := range subs {
		go func(sub *shell.PubSubSubscription) {
			for {
				msg, err := sub.Next()
				if err != nil {
					fmt.Println("Error reading from PubSub:", err)
					return
				}
				msgs, err := message.MessagesFromIPFS(string(msg.Data))
				if err != nil {
					fmt.Println("Error reading from IPFS:", err)
					return
				}
				LocalMessages.AddMany(msgs)
			}
		}(sub)
	}
}

// GetDatabase checks if DB is already set and opened, if not it Sets the database first
func GetDatabase() *sql.DB {
	if DB == nil {
		fmt.Println("Database not set, setting database...")
		SetDatabase()
	}
	return DB
}

func GetMessagesFromDatabase(db *sql.DB) *message.Messages {
	// Get all messages from the database
	rows, err := db.Query("SELECT hash, message, nonce, timestamp FROM messages")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	// Create a new Messages object
	msgs := message.Messages{}
	// Loop through all messages
	fmt.Println("Getting messages from database...")
	for rows.Next() {
		var hash, msg string
		var nonce int
		var timestamp int64
		// Get the values from the database
		err := rows.Scan(&hash, &msg, &nonce, &timestamp)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Got message from database:", hash)

		// Create a new message object
		m := message.Message{
			Message:   msg,
			Nonce:     nonce,
			Timestamp: timestamp,
		}
		// Add the message to the Messages object
		fmt.Println("Adding message to Messages object...")
		msgs.Add(&m)
	}
	return &msgs
}

func InitDatabase(db *sql.DB) {
	var err error
	// Create the table "messages"
	_, err = db.Exec("CREATE TABLE messages(hash TEXT PRIMARY KEY, message TEXT, nonce INTEGER, timestamp INTEGER)")
	if err != nil {
		fmt.Println(err)
	}
	// Create the table "followed_tags"
	_, err = db.Exec("CREATE TABLE followed_tags(tag TEXT)")
	if err != nil {
		fmt.Println(err)
	}
}

// SetDatabase configures DB to be the database to use
// The name used is DatabasePath, but the user will be asked if this correct or if they want to change it
// If the database is already set, it will ask the user if they want to overwrite it
// If the database is not set, it will ask the user if they want to create it
// If the database is set but it doesn't contain the tables "messages" and "followed_tags",
// it will create them
func SetDatabase() {
	// First check if the user is okay with the database path
	fmt.Println("Database path: ", DatabasePath)
	fmt.Println("Is this correct? (y/n)")
	answer := Readline()
	if answer == "n" {
		fmt.Println("Enter the database path: ")
		DatabasePath = Readline()
	}
	// Open the database
	db, err := sql.Open("sqlite", DatabasePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Check if the database is already set
	if DB != nil {
		fmt.Println("Database already set, overwrite? (y/n)")
		answer := Readline()
		if answer == "n" {
			return
		}
	}
	// Check if the database exists
	if _, err := os.Stat(DatabasePath); os.IsNotExist(err) {
		fmt.Println("Database does not exist, create? (y/n)")
		answer := Readline()
		if answer == "n" {
			return
		}
	}
	// Set the database
	DB = db
	// Check if the database contains the tables "messages" and "followed_tags"
	// If not, create them
	InitDatabase(DB)
}
