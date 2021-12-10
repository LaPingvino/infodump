// Infodump is a commandline social network that works over IPFS and communicates through PubSub
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.kiefte.eu/lapingvino/infodump/message"
)

// pubsub "github.com/libp2p/go-libp2p-pubsub"
// ipfs "github.com/ipfs/go-ipfs-api"

// Main function: get the minimal Proof of Work for a message as the first argument, the message as the rest of the argument and return the stamp
func main() {
	// Check if the user provided a number and a message
	if len(os.Args) < 3 {
		fmt.Println("Please provide a number and a message")
		os.Exit(1)
	}
	// Get the number of initial zeroes from the user
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Please provide a number")
		os.Exit(1)
	}
	// Get the message from the user
	msg := strings.Join(os.Args[2:], " ")
	// Create a new message
	m := message.New(msg, n)
	// Print the stamp
	fmt.Println(m.Stamp())
	// Print the message, nonce and number of leading zeroes
	fmt.Println(m.Message, m.Nonce, m.Lead())
}
