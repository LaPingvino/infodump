// Infodump is a commandline social network that works over IPFS and communicates through PubSub
package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"

	"git.kiefte.eu/lapingvino/infodump/message"
	_ "modernc.org/sqlite"
)

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
