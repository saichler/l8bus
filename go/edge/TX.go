package edge

import (
	"errors"
	"github.com/saichler/my.simple/go/utils/logs"
	"github.com/saichler/overlayK8s/go/protocol"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

// loop of Writing data to socket
func (edge *EdgeImpl) writeToSocket() {
	// As long ad the port is active
	for edge.active {
		// Get next data to write to the socket from the TX queue, if no data, this is a blocking call
		data := edge.tx.Next()
		// if the data is not nil
		if data != nil && edge.active {
			//Write the data to the socket
			err := common.Write(data, edge.conn)
			// If there is an error
			if err != nil {
				// If this is not a port on the switch, then try to reconnect.
				if !edge.isSwitch {
					edge.attemptToReconnect()
					err = common.Write(data, edge.conn)
				} else {
					break
				}
			}
		} else {
			// if the data is nil, break and cleanup
			break
		}
	}
	logs.Debug("Connection Write for ", edge.Name(), " ended.")
	edge.Shutdown()
}

// Send Add the raw data to the tx queue to be written to the socket
func (edge *EdgeImpl) Send(data []byte) error {
	// if the port is still active
	if edge.active {
		// Add the data to the TX queue
		edge.tx.Add(data)
	} else {
		return errors.New("Port is not active")
	}
	return nil
}

// Do is wrapping a protobuf with a secure message and send it to the switch
func (edge *EdgeImpl) Do(request *types.Request, destination string, pb proto.Message) error {
	// Create message payload
	data, err := protocol.CreateMessageFor(types.Priority_P0, request, edge.uuid, destination, pb)
	if err != nil {
		logs.Error("Failed to create message:", err)
		return err
	}
	//Send the secure message to the switch
	edge.Send(data)
	return nil
}
