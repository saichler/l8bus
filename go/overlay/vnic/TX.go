package vnic

import (
	"errors"
	"github.com/saichler/shared/go/share/nets"
	"github.com/saichler/shared/go/share/queues"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

type TX struct {
	vnic         *VirtualNetworkInterface
	shuttingDown bool
	// The incoming data queue
	tx *queues.ByteSliceQueue
}

func newTX(vnic *VirtualNetworkInterface) *TX {
	tx := &TX{}
	tx.vnic = vnic
	tx.tx = queues.NewByteSliceQueue("TX", int(vnic.resources.Config().TxQueueSize))
	return tx
}

func (tx *TX) start() {
	go tx.writeToSocket()
}

func (tx *TX) shutdown() {
	tx.shuttingDown = true
	if tx.vnic.conn != nil {
		tx.vnic.conn.Close()
	}
	tx.tx.Shutdown()
}

func (tx *TX) name() string {
	return "TX"
}

// loop of Writing data to socket
func (tx *TX) writeToSocket() {
	// As long ad the port is active
	for tx.vnic.running {
		// Get next data to write to the socket from the TX queue, if no data, this is a blocking call
		data := tx.tx.Next()
		// if the data is not nil
		if data != nil && tx.vnic.running {
			//Write the data to the socket
			err := nets.Write(data, tx.vnic.conn, tx.vnic.resources.Config())
			// If there is an error
			if err != nil {
				if tx.vnic.IsSwitch {
					break
				}
				// If this is not a port on the switch, then try to reconnect.
				if !tx.shuttingDown {
					tx.vnic.reconnect()
					err = nets.Write(data, tx.vnic.conn, tx.vnic.resources.Config())
				} else {
					break
				}
			}
		} else {
			// if the data is nil, break and cleanup
			break
		}
	}
	tx.vnic.resources.Logger().Debug("Connection Write for ", tx.vnic.name, " ended.")
	tx.vnic.Shutdown()
}

// Send Add the raw data to the tx queue to be written to the socket
func (tx *TX) Send(data []byte) error {
	// if the port is still active
	if tx.vnic.running {
		// Add the data to the TX queue
		tx.tx.Add(data)
	} else {
		return errors.New("Port is not active")
	}
	return nil
}

// Do is wrapping a protobuf with a secure message and send it to the switch
func (tx *TX) Do(action types.Action, destination string, pb proto.Message) error {
	// Create message payload
	data, err := tx.vnic.protocol.CreateMessageFor(types.Priority_P0, action, tx.vnic.resources.Config().Local_Uuid,
		tx.vnic.resources.Config().RemoteUuid, destination, pb)
	if err != nil {
		tx.vnic.resources.Logger().Error("Failed to create message:", err)
		return err
	}
	//Send the secure message to the switch
	return tx.Send(data)
}
