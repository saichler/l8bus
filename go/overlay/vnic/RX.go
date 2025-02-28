package vnic

import (
	"github.com/saichler/shared/go/share/nets"
	"github.com/saichler/shared/go/share/queues"
	"github.com/saichler/shared/go/types"
)

type RX struct {
	vnic         *VirtualNetworkInterface
	shuttingDown bool
	// The incoming data queue
	rx *queues.ByteSliceQueue
}

func newRX(vnic *VirtualNetworkInterface) *RX {
	rx := &RX{}
	rx.vnic = vnic
	rx.rx = queues.NewByteSliceQueue("RX", int(vnic.resources.Config().RxQueueSize))
	return rx
}

func (rx *RX) start() {
	go rx.readFromSocket()
	go rx.notifyRawDataListener()
}

func (rx *RX) shutdown() {
	rx.shuttingDown = true
	if rx.vnic.conn != nil {
		rx.vnic.conn.Close()
	}
	rx.rx.Shutdown()
}

func (rx *RX) name() string {
	return "RX"
}

// loop and Read incoming data from the socket
func (rx *RX) readFromSocket() {
	// While the port is active
	for rx.vnic.running {
		// read data ([]byte) from socket
		data, err := nets.Read(rx.vnic.conn, rx.vnic.resources.Config())
		//If therer is an error
		if err != nil {
			if rx.vnic.IsVNet {
				break
			}
			if !rx.shuttingDown {
				rx.vnic.reconnect()
				continue
			} else {
				break
			}
		}
		if data != nil {
			// If still active, write the data to the RX queue
			if rx.vnic.running {
				rx.rx.Add(data)
			}
		} else {
			// If data is nil, it means the port was shutdown
			// so break and cleanup
			break
		}
	}
	rx.vnic.resources.Logger().Debug("RX for ", rx.vnic.name, " ended.")
	//Just in case, mark the port as shutdown so other thread will stop as well.
	rx.vnic.Shutdown()
}

// Notify the RawDataListener on new data
func (rx *RX) notifyRawDataListener() {
	// While the port is active
	for rx.vnic.running {
		// Read next data ([]byte) block
		data := rx.rx.Next()
		// If data is not nil
		if data != nil {
			rx.vnic.stats.RxMsgCount++
			rx.vnic.stats.RxDataCont += int64(len(data))
			// if there is a dataListener, this is a switch
			if rx.vnic.resources.DataListener() != nil {
				rx.vnic.resources.DataListener().HandleData(data, rx.vnic)
			} else {
				msg, err := rx.vnic.protocol.MessageOf(data)
				if err != nil {
					rx.vnic.resources.Logger().Error(err)
					continue
				}
				pb, err := rx.vnic.protocol.ProtoOf(msg)
				if err != nil {
					rx.vnic.resources.Logger().Error(err)
					continue
				}
				// Otherwise call the handler per the action & the type
				if msg.Action == types.Action_Reply {
					request := rx.vnic.requests.getRequest(msg.Sequence)
					request.response = pb
					request.cond.Broadcast()
				} else if msg.Action == types.Action_Notify {
					rx.vnic.resources.ServicePoints().Notify(pb, msg.Action, rx.vnic, msg)
				} else {
					resp, err := rx.vnic.resources.ServicePoints().Handle(pb, msg.Action, rx.vnic, msg)
					if err != nil {
						rx.vnic.resources.Logger().Error(err)
					}
					if msg.IsRequest {
						msg.Action = types.Action_Reply
						msg.Topic = msg.SourceUuid
						msg.SourceUuid = rx.vnic.resources.Config().LocalUuid
						msg.SourceVnetUuid = rx.vnic.resources.Config().RemoteUuid
						msg.IsRequest = false
						msg.IsReply = true
						data, err := rx.vnic.protocol.CreateMessageForm(msg, resp)
						if err != nil {
							rx.vnic.resources.Logger().Error(err)
							return
						}
						rx.vnic.SendMessage(data)
					}
				}
			}
		}
	}
	rx.vnic.resources.Logger().Debug("ND for ", rx.vnic.name, " has Ended")
	rx.vnic.Shutdown()
}
