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
	ipfs "github.com/ipfs/go-ipfs-api"
	_ "modernc.org/sqlite"
)

var LocalMessages = message.Messages{}
var DatabasePath = "infodump.db"
var DB *sql.DB

// Readline reads from a buffered stdin and returns the line
func Readline() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return line
}

// MenuElements is a list of options for the menu, consisting of a description and a function
type MenuElements struct {
	Description string
	Function    func()
}

// Menu presents a menu to the user
func Menu(menu []MenuElements) {
	// Present the user with a menu
	fmt.Println("Welcome to Infodump")
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
		// Present the user with a menu
		Menu([]MenuElements{
			{"Read Messages", ReadMessages},
			{"Write Message", WriteMessage},
			{"Sync Messages", SyncMessages},
			{"Set IPFS Gateway", SetIPFSGateway},
			{"Set Database", SetDatabase},
			{"Configure Followed Tags", ConfigureFollowedTags},
			{"Quit", func() { os.Exit(0) }},
		})

		// After the user has selected an option and before the user gets to chose again, ask to press enter to continue
		fmt.Println("Press enter to continue...")
		Readline()
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

func ReadMessages() {
	// TODO: Read messages from PubSub topic "OLN"
	fmt.Println("Reading messages...")
}

func WriteMessage() {
	// Get a message and an urgency from the user.
	// The urgency is used to set the strength of the Proof of Work
	fmt.Println("Enter a message: ")
	m := Readline()
	fmt.Println("Enter an urgency (higher is stronger but takes longer to produce): ")
	var urgency int
	fmt.Scanln(&urgency)
	// Create a Message object
	msg := message.New(m, urgency, time.Now().Unix())
	// Add the message to LocalMessages
	LocalMessages.Add(msg)
}

func SyncMessages() {
	// Update the database with the messages in LocalMessages
	db := GetDatabase()
	// Put the messages in the database
	for k, m := range LocalMessages {
		_, err := db.Exec("INSERT INTO messages(hash, message, nonce, timestamp) VALUES(?,?,?,?)", m.Stamp(), m.Message, m.Nonce, m.Timestamp)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Message", k, "synced")
		}
	}

	fmt.Println("Now loading messages from database...")
	// Get the messages from the database in a new Messages object
	var messages = message.Messages{}
	rows, err := db.Query("SELECT * FROM messages")
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		var hash, msg string
		var nonce int
		var timestamp int64
		fmt.Println("Reading message...")
		err := rows.Scan(&hash, &msg, &nonce, &timestamp)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Message", hash, "loaded")

		m := message.Message{
			Message:   msg,
			Nonce:     nonce,
			Timestamp: timestamp,
		}

		messages[hash] = &m
	}
	fmt.Println("Messages loaded from database")
	// Add the messages to the IPFS network
	cid, err := messages.AddToIPFS()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Messages synced to IPFS: ", cid)
	}
}

func TestIPFSGateway(gateway string) error {
	// Test the IPFS gateway
	fmt.Println("Testing IPFS gateway...")
	ipfs := ipfs.NewShell(gateway)
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
	if strings.HasPrefix(answer, "n") {
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
		if strings.HasPrefix(answer, "n") {
			return
		}
	}
	// Check if the database exists
	if _, err := os.Stat(DatabasePath); os.IsNotExist(err) {
		fmt.Println("Database does not exist, create? (y/n)")
		answer := Readline()
		if strings.HasPrefix(answer, "n") {
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
	fmt.Println("Enter the tags you want to follow, separated by spaces: ")
	tags := Readline()
	// Split the tags into an array and insert them into database DB
	// using table "followed_tags"
	tagArray := strings.Split(tags, " ")
	for _, tag := range tagArray {
		_, err := db.Exec("INSERT INTO followed_tags(tag) VALUES(?)", tag)
		if err != nil {
			fmt.Println(err)
		}
	}
}
