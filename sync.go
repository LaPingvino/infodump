package main

import (
	"fmt"

	"git.kiefte.eu/lapingvino/infodump/message"
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
