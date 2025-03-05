package vnic

import (
	"errors"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/shared/go/share/nets"
	"github.com/saichler/shared/go/share/queues"
	"github.com/saichler/shared/go/types"
	"strconv"
	"time"
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

func (this *TX) start() {
	go this.writeToSocket()
}

func (this *TX) shutdown() {
	this.shuttingDown = true
	if this.vnic.conn != nil {
		this.vnic.conn.Close()
	}
	this.tx.Shutdown()
}

func (this *TX) name() string {
	return "TX"
}

// loop of Writing data to socket
func (this *TX) writeToSocket() {
	// As long ad the port is active
	for this.vnic.running {
		// Get next data to write to the socket from the TX queue, if no data, this is a blocking call
		data := this.tx.Next()
		// if the data is not nil
		if data != nil && this.vnic.running {
			//Write the data to the socket
			err := nets.Write(data, this.vnic.conn, this.vnic.resources.Config())
			// If there is an error
			if err != nil {
				if this.vnic.IsVNet {
					break
				}
				// If this is not a port on the switch, then try to reconnect.
				if !this.shuttingDown && this.vnic.running {
					this.vnic.reconnect()
					err = nets.Write(data, this.vnic.conn, this.vnic.resources.Config())
				} else {
					break
				}
			}
			this.vnic.stats.LastMsgTime = time.Now().UnixMilli()
			this.vnic.stats.TxMsgCount++
			this.vnic.stats.TxDataCount += int64(len(data))
		} else {
			// if the data is nil, break and cleanup
			break
		}
	}
	this.vnic.resources.Logger().Debug("TX for ", this.vnic.name, " ended.")
	this.vnic.Shutdown()
}

// Send Add the raw data to the tx queue to be written to the socket
func (this *TX) SendMessage(data []byte) error {
	// if the port is still active
	if this.vnic.running {
		// Add the data to the TX queue
		this.tx.Add(data)
	} else {
		return errors.New("Port is not active")
	}
	return nil
}

// Unicast is wrapping a protobuf with a secure message and send it to the vnet
func (this *TX) Unicast(action types.Action, destination string, any interface{}, p types.Priority, isRequest, isReply bool, msgNum int32, tr *types.Transaction) error {
	if len(destination) != protocol.UNICAST_ADDRESS_SIZE {
		return errors.New("Invalid destination address " + destination + " size " + strconv.Itoa(len(destination)))
	}
	return this.Multicast(action, 0, destination, any, p, isRequest, isReply, msgNum, tr)
}

// Multicast is wrapping a protobuf with a secure message and send it to the vnet topic
func (this *TX) Multicast(action types.Action, area int32, topic string, any interface{}, p types.Priority, isRequest, isReply bool, msgNum int32, tr *types.Transaction) error {
	// Create message payload
	data, err := this.vnic.protocol.CreateMessageFor(area, topic, p, action,
		this.vnic.resources.Config().LocalUuid, this.vnic.resources.Config().RemoteUuid, any, isRequest, isReply, msgNum, tr)
	if err != nil {
		this.vnic.resources.Logger().Error("Failed to create message:", err)
		return err
	}
	//Send the secure message to the vnet
	return this.SendMessage(data)
}
