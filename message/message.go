package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"sort"
	"sync"
	"time"

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
	Message   string
	Timestamp int64
	Nonce     int
}

// String method for Message: "Message *hash* sent at *human readable timestamp* with nonce *nonce*:\n*message*"
func (m *Message) String() string {
	return fmt.Sprintf("Message %x sent at %s with nonce %d:\n%s", m.Hash(), time.Unix(m.Timestamp, 0).Format(time.RFC3339), m.Nonce, m.Message)
}

// SortNum of a Message returns a number that can be used to sort messages by importance
// The number is calculated by taking the timestamp of the message and
// adding an importance factor of 2^(leading zeros of hash / 8) to it
// If the timestamp is in the future, return 0 instead so the message will be discarded unless there are almost no messages
func (m *Message) SortNum() int64 {
	if m.Timestamp > time.Now().Unix() {
		return 0
	}
	importance := math.Pow(2, float64(CountLeadingZeroes(m.Hash()))/8)
	return m.Timestamp + int64(importance)
}

// MessageList returns a slice of Messages sorted by importance
// The slice is sorted by the SortNum method of the Message type
// It can be created from a Messages map by calling the Messages.Sorted method
func (m *Messages) MessageList() []*Message {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var msgs []*Message
	for _, msg := range m.msgs {
		msgs = append(msgs, msg)
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].SortNum() > msgs[j].SortNum()
	})
	return msgs
}

// Proof of Work: Find the nonce for a message by hashing the message and checking for at least n initial zeroes in the binary representation of the resulting hash
// If it takes too long, return an error
func (msg *Message) ProofOfWork(n int, timeout time.Duration) error {
	// Create a local copy of the message and start counting
	m := *msg
	m.Nonce = 0
	start := time.Now()
	// Loop until we find a nonce that satisfies the proof of work
	// If the nonce is not found within the timeout, return an error
	for {
		// Increment the nonce and hash the message
		m.Nonce++
		hash := m.Hash()
		// If the hash has at least n initial zeroes, we have found a valid nonce
		if CountLeadingZeroes(hash) >= n {
			break
		}
		// If the nonce is not found within the timeout, return an error
		if time.Since(start) > timeout {
			return fmt.Errorf("proof of work timed out")
		}
	}
	// Set the message to the local copy
	*msg = m
	return nil
}

// Get the SHA256 hash of a message plus the timestamp plus the nonce as a byte slice
func (m *Message) Hash() [32]byte {
	hash := sha256.Sum256([]byte(m.Message + fmt.Sprintf("%d", m.Timestamp) + fmt.Sprintf("%d", m.Nonce)))
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

func New(msg string, n int, timestamp int64, timeout time.Duration) (*Message, error) {
	m := Message{Message: msg, Timestamp: timestamp}
	err := m.ProofOfWork(n, timeout)
	return &m, err
}

// Map Messages maps the stamp of the message to the message itself
type Messages struct {
	lock sync.RWMutex
	msgs map[string]*Message
}

// MessagesFromIPFS takes a CID and returns a Messages map
func MessagesFromIPFS(cid string) (*Messages, error) {
	emptyMsgs := Messages{msgs: make(map[string]*Message)}
	// Create an IPFS instance based on the IPFSGateway
	myIPFS := ipfs.NewShell(IPFSGateway)
	// Get the JSON from IPFS
	jsonr, err := myIPFS.Cat(cid)
	if err != nil {
		return &emptyMsgs, err
	}
	jsonb, err := io.ReadAll(jsonr)
	if err != nil {
		return &emptyMsgs, err
	}
	// Unmarshal the JSON into a Messages map
	var messages Messages
	err = json.Unmarshal(jsonb, &(messages.msgs))
	if err != nil {
		return &emptyMsgs, err
	}
	return &messages, nil
}

// Trim the Messages map to the given number of messages based on the importance of the messages
func (m *Messages) Trim(n int) {
	// Create a slice of Messages sorted by importance
	// Cannot lock the Messages map yet, Messages.MessageList() will lock it
	msgs := m.MessageList()
	// Now lock the Messages map and trim the map to the given number of messages
	m.lock.Lock()
	defer m.lock.Unlock()
	// If the number of messages is less than or equal to the number of messages to keep, do nothing
	if len(msgs) <= n {
		return
	}
	fmt.Println("Trimming messages")
	// Otherwise, remove the messages after the nth message from the map
	for i := n; i < len(msgs); i++ {
		delete(m.msgs, msgs[i].Stamp())
	}
}

// Add a message to the Messages map
func (m *Messages) Add(msg *Message) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.msgs == nil {
		m.msgs = make(map[string]*Message)
	}
	m.msgs[msg.Stamp()] = msg
}

// AddMany adds another Messages map to the current Messages map
func (m *Messages) AddMany(msgs *Messages) {
	m.lock.Lock()
	defer m.lock.Unlock()
	msgs.lock.RLock()
	defer msgs.lock.RUnlock()
	if m.msgs == nil {
		m.msgs = make(map[string]*Message)
	}
	for _, msg := range msgs.msgs {
		m.msgs[msg.Stamp()] = msg
	}
}

// Remove a message from the Messages map by stamp
func (m *Messages) Remove(stamp string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.msgs, stamp)
}

// Return a JSON representation of the Messages map
func (m *Messages) JSON() ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return json.Marshal(m.msgs)
}

// Do something with each message in the Messages map
func (m *Messages) Each(f func(msg *Message)) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.msgs == nil {
		return
	}
	for _, msg := range m.msgs {
		f(msg)
	}
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
	// Turn the JSON into a Reader and add it to IPFS
	cid, err := myIPFS.Add(bytes.NewReader(json))
	if err != nil {
		return "", err
	}
	return cid, nil
}
