package main

import (
	"fmt"

	"git.kiefte.eu/lapingvino/infodump/message"
	shell "github.com/ipfs/go-ipfs-api"
)

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
