package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/bits"

	ipfs "github.com/ipfs/go-ipfs-api"
)

var IPFSGateway = "http://localhost:5001" // IPFS Gateway used to connect to the IPFS daemon

// Set up an IPFS instance based on the IPFSGateway
func InitIPFS() *ipfs.Shell {
	return ipfs.NewShell(IPFSGateway)
}

// Messages on Infodump use a "stamp" using the hashcash algorithm to prevent spam and enable storing messages by importance
// The Message type contains the message itself and a nonce that is used to verify the stamp
type Message struct {
	Message string
	Nonce   int
}

// Proof of Work: Find the nonce for a message by hashing the message and checking for at least n initial zeroes in the binary representation of the resulting hash
func (msg *Message) ProofOfWork(n int) {
	// Create a local copy of the message and start counting
	m := *msg
	m.Nonce = 0
	// Loop until we find a nonce that satisfies the proof of work
	for {
		// Increment the nonce and hash the message
		m.Nonce++
		hash := m.Hash()
		// If the hash has at least n initial zeroes, we have found a valid nonce
		if CountLeadingZeroes(hash) >= n {
			break
		}
	}
	// Set the message to the local copy
	*msg = m
}

// Get the SHA256 hash of a message plus the nonce as a byte slice
func (m *Message) Hash() [32]byte {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d", m.Message, m.Nonce)))
	return hash
}

// Bitwise count leading zeroes in a byte slice
func CountLeadingZeroes(b [32]byte) int {
	count := 0
	for _, v := range b {
		// If the byte is zero, that means there are 8 leading zeroes and we need to continue counting
		if v == 0 {
			count += 8
		} else { // Otherwise, we add the number of leading zeroes in the byte and stop counting
			// Bitwise count leading zeroes in the byte
			count += bits.LeadingZeros8(v)
			// If the byte is not zero, we can stop counting
			break
		}
	}
	return count
}

// Get the stamp of the message
func (m *Message) Stamp() string {
	return fmt.Sprintf("%x", m.Hash())
}

// Lead is a method that returns the number of leading zeroes in the hash of a message plus its nonce
func (m *Message) Lead() int {
	return CountLeadingZeroes(m.Hash())
}

func New(msg string, n int) *Message {
	m := Message{Message: msg}
	m.ProofOfWork(n)
	return &m
}

// Map Messages maps the stamp of the message to the message itself
type Messages map[string]*Message

// Add a message to the Messages map
func (m *Messages) Add(msg *Message) {
	(*m)[msg.Stamp()] = msg
}

// Remove a message from the Messages map by stamp
func (m *Messages) Remove(stamp string) {
	delete(*m, stamp)
}

// Return a JSON representation of the Messages map
func (m *Messages) JSON() ([]byte, error) {
	return json.Marshal(m)
}

// Add the messages as a JSON object to IPFS
func (m *Messages) AddToIPFS() (string, error) {
	json, err := m.JSON()
	if err != nil {
		return "", err
	}
	return AddJSONToIPFS(json)
}

func AddJSONToIPFS(json []byte) (string, error) {
	// Create an IPFS instance based on the IPFSGateway
	myIPFS := ipfs.NewShell(IPFSGateway)
	// Turn the JSON into a DAG node and return the CID
	cid, err := myIPFS.Add(bytes.NewReader(json))
	if err != nil {
		return "", err
	}
	return cid, nil
}
