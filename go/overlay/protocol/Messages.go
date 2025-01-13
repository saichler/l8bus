package protocol

import (
	"github.com/saichler/shared/go/types"
)

func CreateHeader(msg *types.Message) []byte {
	header := make([]byte, 109)
	for i, c := range msg.SourceUuid {
		header[i] = byte(c)
	}
	for i, c := range msg.SourceSwitchUuid {
		header[i+36] = byte(c)
	}
	for i, c := range msg.Destination {
		header[i+72] = byte(c)
	}
	header[108] = byte(msg.Priority)
	return header
}

func HeaderOf(data []byte) (string, string, string, types.Priority) {
	// Source will always be Uuid
	source := string(data[0:36])
	// Source switch will always be Uuid
	sourceSwitch := string(data[36:72])
	//Destination, either than being a uuid can also be a topic/multicast so it might not be full 16 bytes
	dest := make([]byte, 0)
	for i := 72; i < 108; i++ {
		if data[i] == 0 {
			break
		}
		dest = append(dest, data[i])
	}
	pri := types.Priority(data[108])
	return source, sourceSwitch, string(dest), pri
}
