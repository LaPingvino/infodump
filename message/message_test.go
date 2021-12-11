// Test the message package
package message_test

import (
	"testing"
	"time"

	"git.kiefte.eu/lapingvino/infodump/message"
)

// Test if the code finishes in time
func TestMessageTimeout(t *testing.T) {
	time.AfterFunc(10*time.Second, func() {
		t.Error("Test timed out")
	})
	_ = message.New("Hello World!", 16, time.Now().Unix())
	// Success
}
