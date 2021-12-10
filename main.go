// Infodump is a commandline social network that works over IPFS and communicates through PubSub
package main

import (
	"bufio"
	"fmt"
	"os"

	"git.kiefte.eu/lapingvino/infodump/message"
	_ "modernc.org/sqlite"
)

var LocalMessages = message.Messages{}

// Readline reads from a buffered stdin and returns the line
func Readline() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return line
}

func main() {
	// Set message IPFS client to use the local IPFS node
	message.IPFSGateway = "http://localhost:5001"

	// Run a loop and present a menu to the user to read messages, write messages or quit the program
	for {
		// Present the user with a menu
		fmt.Println("Welcome to Infodump")
		fmt.Println("1. Read Messages")
		fmt.Println("2. Write Messages")
		fmt.Println("3. Sync messages")
		fmt.Println("4. Quit")
		fmt.Println("Enter your choice: ")
		var choice int
		fmt.Scanln(&choice)
		switch choice {
		case 1:
			// Read the messages from PubSub topic "OLN"
			fmt.Println("Reading messages...")
			// TODO: Read messages from PubSub topic "OLN"
		case 2:
			// Write Messages
			fmt.Println("Writing Messages")
			// Get a message and an urgency from the user.
			// The urgency is used to set the strength of the Proof of Work
			fmt.Println("Enter a message: ")
			m := Readline()
			fmt.Println("Enter an urgency (higher is stronger but takes longer to produce): ")
			var urgency int
			fmt.Scanln(&urgency)
			// Create a Message object
			msg := message.New(m, urgency)
			// Add the message to LocalMessages
			LocalMessages.Add(msg)
		case 3:
			// Sync messages
			fmt.Println("Syncing messages...")
			// Upload the LocalMessages to IPFS
			hash, err := LocalMessages.AddToIPFS()
			if err != nil {
				fmt.Println(err)
			}
			// TODO: Publish the hash to the PubSub topic "OLN"
			fmt.Println("Hash: ", hash)
		case 4:
			// Quit
			fmt.Println("Quitting")
			os.Exit(0)
		default:
			fmt.Println("Invalid Choice")
		}
	}
}
