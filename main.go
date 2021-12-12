// Infodump is a commandline social network that works over IPFS and communicates through PubSub
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"git.kiefte.eu/lapingvino/infodump/message"
	shell "github.com/ipfs/go-ipfs-api"
	_ "modernc.org/sqlite"
)

var LocalMessages = message.Messages{}
var DatabasePath = "infodump.db"
var DB *sql.DB
var MessageCache string

// Readline reads from a buffered stdin and returns the line
func Readline() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return line[:len(line)-1]
}

// MenuElements is a list of options for the menu, consisting of a description and a function
type MenuElements struct {
	Description string
	Function    func()
}

// Menu presents a menu to the user
func Menu(menu []MenuElements) {
	// Present the user with a menu
	for i, e := range menu {
		fmt.Println(i+1, e.Description)
	}
	fmt.Println("Enter your choice: ")
	var choice int
	fmt.Scanln(&choice)
	if choice > 0 && choice <= len(menu) {
		menu[choice-1].Function()
	} else {
		fmt.Println("Invalid Choice")
	}
}

func main() {
	fmt.Println("Welcome to Infodump")

	// Set message IPFS client to use the local IPFS node
	message.IPFSGateway = "http://localhost:5001"

	// If the first argument to the command is a valid link, use that for the IPFSGateway instead
	if len(os.Args) > 1 {
		u, err := url.Parse(os.Args[1])
		if err == nil {
			message.IPFSGateway = u.String()
		}
	}

	err := TestIPFSGateway(message.IPFSGateway)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Run a loop and present a menu to the user to
	// read messages
	// write messages
	// sync messages
	// set the IPFS gateway
	// set the database
	// configure the followed tags
	// quit the program
	for {
		// Check if the database is set, if so, show the followed tags
		if DB != nil {
			fmt.Println("Following tags:")
			for _, tag := range GetFollowedTags(DB) {
				fmt.Print(tag, " ")
				fmt.Println()
			}
		}
		// Present the user with a menu
		Menu([]MenuElements{
			{"Start OLN Listener", StartOLNListener},
			{"Read Messages", ReadMessages},
			{"Write Message", WriteMessage},
			{"Sync Messages", SyncMenu},
			{"Settings", SettingsMenu},
			{"Quit", func() { os.Exit(0) }},
		})

		// After the user has selected an option and before the user gets to chose again, ask to press enter to continue
		fmt.Println("Press enter to continue...")
		Readline()
	}
}

// A submenu for all settings
func SettingsMenu() {
	// Present the user with a menu
	Menu([]MenuElements{
		{"Set IPFS Gateway", SetIPFSGateway},
		{"Set Database", SetDatabase},
		{"Configure Followed Tags", ConfigureFollowedTags},
		{"Back", func() {}},
	})
}

// A menu for the several parts of Sync:
// saving messages to the local database
// reading messages from the local database
// reading messages from the network
// writing messages to the network
func SyncMenu() {
	// Present the user with a menu
	Menu([]MenuElements{
		{"Save Messages to Database", SaveMessagesToDatabase},
		{"Read Messages from Database", ReadMessagesFromDatabase},
		{"Read Messages from Network", ReadMessagesFromNetwork},
		{"Write Messages to Network", WriteMessagesToNetwork},
		{"Back", func() {}},
	})
}

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
					fmt.Println(err)
					return
				}
				msgs, err := message.MessagesFromIPFS(string(msg.Data))
				if err != nil {
					fmt.Println(err)
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

// ReadMessages shows the messages in LocalMessages sorted by importance, 10 at a time
func ReadMessages() {
	// Use LocalMessages to get the messages and get the sorted list of messages
	// through the MessageList method
	msgs := LocalMessages.MessageList()
	// Loop through the messages and print them
	for i, m := range msgs {
		fmt.Println(m)
		if i%10 == 9 {
			fmt.Println("Press enter to continue... Type anything to stop")
			contp := Readline()
			if contp != "" {
				return
			}
		}
	}
}

func WriteMessage() {
	// Get a message and an urgency from the user.
	// The urgency is used to set the strength of the Proof of Work
	// If there is a message in the MessageCache, ask the user if they want to use it
	// If there is no message in the MessageCache, ask the user to write a message
	var m string
	if MessageCache != "" {
		fmt.Println("You wrote a message before that you didn't send yet. Do you want to use that message?")
		fmt.Println("The message is:", MessageCache)
		fmt.Println("Type 'yes' to use the message, or anything else to write a new message")
		usep := Readline()
		if usep == "yes" {
			fmt.Println("Using message from cache")
			m = MessageCache
		} else {
			fmt.Println("Writing new message")
			m = usep
		}
	} else {
		fmt.Println("Write a message:")
		m = Readline()
	}
	fmt.Println("Enter an urgency (higher is stronger but takes longer to produce): ")
	var urgency int
	fmt.Scanln(&urgency)
	fmt.Println("How many seconds should we wait for the POW to be done? (default is 5): ")
	var powtime int
	fmt.Scanln(&powtime)
	if powtime == 0 {
		powtime = 5
	}
	// Create a new message object
	msg, err := message.New(m, urgency, time.Now().Unix(), time.Duration(powtime)*time.Second)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Want to try again with another urgency or timeout? (y/n)")
		contp := Readline()
		if contp == "y" {
			MessageCache = m
			WriteMessage()
		}
		return
	}
	// MessageCache can be discarded after this point
	MessageCache = ""
	// Add the message to LocalMessages
	LocalMessages.Add(msg)

	// Ask if the user wants to write another message or save the messages to the database
	fmt.Println("Do you want to write another message? (y/n)")
	contp := Readline()
	if contp == "y" {
		WriteMessage()
	} else {
		fmt.Println("Do you want to save the messages to the database? (y/n)")
		contp := Readline()
		if contp == "y" {
			SaveMessagesToDatabase()
		}
	}
}

// Implementing the Sync Menu options
// SaveMessagesToDatabase, ReadMessagesFromDatabase, ReadMessagesFromNetwork and WriteMessagesToNetwork

// SaveMessagesToDatabase saves the messages in LocalMessages to the database
func SaveMessagesToDatabase() {
	// Update the database with the messages in LocalMessages
	db := GetDatabase()
	// Put the messages in the database
	LocalMessages.Each(func(m *message.Message) {
		_, err := db.Exec("INSERT INTO messages(hash, message, nonce, timestamp) VALUES(?,?,?,?)", m.Stamp(), m.Message, m.Nonce, m.Timestamp)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Message", m.Stamp(), "saved to database")
		}
	})
}

// ReadMessagesFromDatabase reads the messages from the database and adds them to LocalMessages
func ReadMessagesFromDatabase() {
	// Get the messages from the database
	messages := GetMessagesFromDatabase(GetDatabase())
	// Add the messages to LocalMessages
	LocalMessages.AddMany(messages)
}

// ReadMessagesFromNetwork reads the messages from the IPFS network and adds them to LocalMessages
func ReadMessagesFromNetwork() {
	// Get the messages from the IPFS network
	// TODO: Implement this
}

// WriteMessagesToNetwork writes the messages in LocalMessages to the IPFS network
func WriteMessagesToNetwork() {
	// Add the messages to the IPFS network
	cid, err := LocalMessages.AddToIPFS()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Messages synced to IPFS: ", cid)
	}
}

func TestIPFSGateway(gateway string) error {
	// Test the IPFS gateway
	fmt.Println("Testing IPFS gateway...")
	ipfs := shell.NewShell(gateway)
	_, err := ipfs.ID()
	return err
}

func SetIPFSGateway() {
	// Set the IPFS gateway
	fmt.Println("Enter the IPFS gateway: ")
	var gateway string
	fmt.Scanln(&gateway)
	// Test the IPFS gateway
	err := TestIPFSGateway(gateway)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("IPFS gateway set to: ", gateway)
		message.IPFSGateway = gateway
	}
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

func ConfigureFollowedTags() {
	db := GetDatabase()
	fmt.Println("At the moment you follow the following tags:")
	// Get the tags that the user is following
	tags := GetFollowedTags(db)
	// Show the tags
	for _, tag := range tags {
		fmt.Print(tag, " ")
	}
	fmt.Println()
	fmt.Println("Enter the tags you want to follow, separated by spaces\nTo remove tags, prefix them with a minus sign: ")
	newtags := Readline()
	// Split the tags into an array and insert them into database DB
	// using table "followed_tags"
	tagArray := strings.Split(newtags, " ")
	for _, tag := range tagArray {
		tag = strings.Trim(tag, " \n")
		if !strings.HasPrefix(tag, "-") {
			_, err := db.Exec("INSERT INTO followed_tags(tag) VALUES(?)", tag)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := db.Exec("DELETE FROM followed_tags WHERE tag=?", tag[1:])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func GetFollowedTags(db *sql.DB) []string {
	// Get the tags from the database
	rows, err := db.Query("SELECT tag FROM followed_tags")
	if err != nil {
		fmt.Println(err)
	}
	var tags []string
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			fmt.Println(err)
		}
		tags = append(tags, tag)
	}
	return tags
}
