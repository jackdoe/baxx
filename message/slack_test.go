package message

import (
	"os"
	"testing"
)

func TestAll(t *testing.T) {
	err := SendSlack(os.Getenv("BAXX_SLACK_PANIC"), "hello", "world")
	if err != nil {
		t.Fatal(err)
	}
}
