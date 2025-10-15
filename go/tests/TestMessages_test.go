package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
)

func TestMessages(t *testing.T) {
	fmt.Println("Message Count 1", protocol.MessagesCreated())
	time.Sleep(time.Second * 5)
	fmt.Println("Message Count 2", protocol.MessagesCreated())
	time.Sleep(time.Second * 5)
	fmt.Println("Message Count 2", protocol.MessagesCreated())
}
