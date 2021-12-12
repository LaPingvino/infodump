// Test the message package
package message_test

import (
	"testing"
	"time"

	"git.kiefte.eu/lapingvino/infodump/message"
)

// Test if creating a proof of work of 16 leading zeros finishes in 10 seconds
func TestMessageTimeout(t *testing.T) {
	msg := message.Message{Message: "test", Timestamp: time.Now().Unix()}
	err := msg.ProofOfWork(16, 10*time.Second)
	if err != nil {
		t.Error(err)
	}
}
