package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
)

func TestMessages(t *testing.T) {
	time.Sleep(time.Second * 15)
	fmt.Println("Message Count 3")
	protocol.MsgLog.Print()
}
