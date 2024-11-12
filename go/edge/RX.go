package edge

import (
	"github.com/saichler/my.simple/go/common"
	"github.com/saichler/my.simple/go/net/protocol"
	"github.com/saichler/my.simple/go/utils/logs"
)

// loop and Read incoming data from the socket
func (edge *EdgeImpl) readFromSocket() {
	// While the port is active
	for edge.active {
		// read data ([]byte) from socket
		data, err := common.Read(edge.conn)

		//If therer is an error
		if err != nil {
			// If this is not a port from the switch side
			if !edge.isSwitch {
				// Attempt to reconnect
				edge.attemptToReconnect()
				// And try to read the data again
				data, err = common.Read(edge.conn)
			} else {
				// If this is the receiving port, break and clean resources.
				logs.Error(err)
				break
			}
		}
		if data != nil {
			// If still active, write the data to the RX queue
			if edge.active {
				edge.rx.Add(data)
			}
		} else {
			// If data is nil, it means the port was shutdown
			// so break and cleanup
			break
		}
	}
	logs.Debug("Connection Read for ", edge.Name(), " ended.")
	//Just in case, mark the port as shutdown so other thread will stop as well.
	edge.Shutdown()
}

// Notify the RawDataListener on new data
func (edge *EdgeImpl) notifyRawDataListener() {
	// While the port is active
	for edge.active {
		// Read next data ([]byte) block
		data := edge.rx.Next()
		// If data is not nil
		if data != nil {
			// if there is a dataListener, this is a switch
			if edge.dataListener != nil {
				edge.dataListener.HandleData(data, edge)
			} else {
				msg, err := protocol.MessageOf(data)
				if err != nil {
					logs.Error(err)
					continue
				}
				pb, err := protocol.ProtoOf(msg, edge.registry)
				if err != nil {
					logs.Error(err)
					continue
				}
				// Otherwise call the handler per the action & the type
				edge.servicePoints.Handle(pb, msg.Action, edge)
			}
		}
	}
	logs.Debug("notify data listener for ", edge.Name(), " Ended")
	edge.Shutdown()
}
