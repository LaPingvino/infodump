package main

import (
	"fmt"
	"time"

	"git.kiefte.eu/lapingvino/infodump/message"
)

var LocalMessages = message.Messages{}
var MessageCache string

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

func TrimMessages() {
	// Get the number of messages to keep from the user
	fmt.Println("How many messages do you want to keep?")
	var keep int
	fmt.Scanln(&keep)
	// Trim the messages
	LocalMessages.Trim(keep)
}
