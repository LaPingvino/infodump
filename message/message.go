package message

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/bits"

	ipfs "github.com/ipfs/go-ipfs-api"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

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

// Datatype Messages that contains a slice of message.Message that is to be stored on IPFS as JSON
type Messages struct {
	Messages []message.Message `json:"messages"`
}

// JSON encodes the Messages struct into a JSON string
func (m *Messages) JSON() string {
	json, _ := json.Marshal(m)
	return string(json)
}

// Push the JSON string to IPFS and return the hash
func (m *Messages) Push(ipfs *ipfs.IpfsApi) (string, error) {
	hash, err := ipfs.BlockPut(m.JSON())
	return hash, err
}

// Read the JSON string from IPFS and return the Messages struct
func (m *Messages) Read(ipfs *ipfs.IpfsApi, hash string) error {
	json, err := ipfs.BlockGet(hash)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(json), &m)
	return err
}

// Save the Messages struct to the database
func (m *Messages) Save(db *sql.DB) error {
	stmt, err := db.Prepare("INSERT INTO messages (hash) VALUES (?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(m.JSON())
	if err != nil {
		return err
	}
	return nil
}

// Push the Messages struct to IPFS and send the hash to the PubSub topic
func (m *Messages) Publish(ipfs *ipfs.IpfsApi, ps *pubsub.PubSub, topic string) error {
	hash, err := m.Push(ipfs)
	if err != nil {
		return err
	}
	err = ps.Publish(topic, []byte(hash))
	return err
}

// ListenAndSave takes an IPFS instance, a PubSub instance, a database and a topic
// Listen on the PubSub topic, look up the hash and read the Messages struct
// On first receipt, save the Messages struct to the database
func ListenAndSave(ipfs *ipfs.IpfsApi, ps *pubsub.PubSub, db *sql.DB, topic string) error {
	// Subscribe to the topic
	sub, err := ps.Subscribe(topic)
	if err != nil {
		return err
	}
	// Listen for messages on the topic
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			return err
		}
		// Read the Messages struct from IPFS
		m := Messages{}
		err = m.Read(ipfs, string(msg.Data))
		if err != nil {
			return err
		}
		// Save the Messages struct to the database
		err = m.Save(db)
		if err != nil {
			return err
		}
	}
}
