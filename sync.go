package main

import (
	"fmt"
	"strings"

	"git.kiefte.eu/lapingvino/infodump/message"
	shell "github.com/ipfs/go-ipfs-api"
)

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
	// Ask for a CID from the user and get the messages from the IPFS network
	// Add the messages to LocalMessages
	fmt.Println("Enter the CID of the messages to read: ")
	cid := Readline()
	messages, err := message.MessagesFromIPFS(cid)
	if err != nil {
		fmt.Println(err)
	} else {
		LocalMessages.AddMany(messages)
	}
}

// WriteMessagesToNetwork writes the messages in LocalMessages to the IPFS network
func WriteMessagesToNetwork() {
	// Create an IPFS shell to later publish the messages to via PubSub
	myIPFS := shell.NewShell(message.IPFSGateway)
	// Add the messages to the IPFS network
	cid, err := LocalMessages.AddToIPFS()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Messages synced to IPFS: ", cid)
	}
	// Publish the CID of LocalMessages to the OLN tag
	err = myIPFS.PubSubPublish("OLN", cid)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Published CID", cid, "to the main network")
	}
	// Define what is considered a tag by regex
	tagDefinition := []string{
		// Hashtags
		"#[a-zA-Z0-9]+",
		// Mentions
		"@[a-zA-Z0-9]+",
		// Links
		"https?://[a-zA-Z0-9./]+",
	}
	// Split each message on spaces and map the resulting strings
	// to the message stamp
	messages := make(map[string][]*message.Message)
	LocalMessages.Each(func(m *message.Message) {
		tags := strings.Split(m.Message, " ")
		itertags:		
		for _, tag := range tags {
			// Check if the tag matches any of the tag definitions
			// if not, continue to the next tag
			for _, regex := range tagDefinition {
				if !strings.Contains(tag, regex) {
					continue itertags
				}
			}
			// If the tag matches any of the tag definitions,
			// add the message to the messages map
			messages[tag] = append(messages[tag], m)
		}
	})
	// Create a new set of messages per tag and send them over PubSub
	// Per message, invoke Add on the tags map entry
	tags := make(map[string]*message.Messages)
	for tag, messages := range messages {
		if tags[tag] == nil {
			tags[tag] = &message.Messages{}
		}
		for _, m := range messages {
			tags[tag].Add(m)
		}
		// Publish the messages to the IPFS network and get a CID
		cid, err := tags[tag].AddToIPFS()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Messages synced to IPFS: ", cid)
		}
		// Publish the messages via PubSub on the IPFS shell
		err = myIPFS.PubSubPublish("oln-"+tag, cid)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Messages published to the rest of the network for tag: ", tag)
		}
	}

}
