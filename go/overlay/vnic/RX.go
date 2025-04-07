package vnic

import (
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/shared/go/share/queues"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/nets"
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
	rx.rx = queues.NewByteSliceQueue("RX", int(vnic.resources.SysConfig().RxQueueSize))
	return rx
}

func (rx *RX) start() {
	go rx.readFromSocket()
	for i := 0; i < 10; i++ {
		go rx.notifyRawDataListener()
	}
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
		data, err := nets.Read(rx.vnic.conn, rx.vnic.resources.SysConfig())
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
				//rx.runRawHandleMessage(data)
				rx.vnic.resources.DataListener().HandleData(data, rx.vnic)
			} else {
				msg, err := rx.vnic.protocol.MessageOf(data)
				if err != nil {
					rx.vnic.resources.Logger().Error(err)
					continue
				}
				pb, err := rx.vnic.protocol.ElementsOf(msg)
				if err != nil {
					rx.vnic.resources.Logger().Error(err)
					if msg.Request() {
						resp := object.NewError(err.Error())
						err = rx.vnic.Reply(msg, resp)
						if err != nil {
							rx.vnic.resources.Logger().Error(err)
						}
					} else if msg.Reply() {
						resp := object.NewError(err.Error())
						request := rx.vnic.requests.getRequest(msg.Sequence(), rx.vnic.resources.SysConfig().LocalUuid)
						request.response = resp
						request.cond.Broadcast()
					}
					continue
				}

				//This is a reply message, should not find a handler
				//and just notify
				if msg.Reply() {
					request := rx.vnic.requests.getRequest(msg.Sequence(), rx.vnic.resources.SysConfig().LocalUuid)
					request.response = pb
					request.cond.Broadcast()
					continue
				}
				// Otherwise call the handler per the action & the type
				rx.handleMessage(msg, pb)
			}
		}
	}
	rx.vnic.resources.Logger().Debug("ND for ", rx.vnic.name, " has Ended")
	rx.vnic.Shutdown()
}

func (rx *RX) handleMessage(msg common.IMessage, pb common.IElements) {
	handler, ok := rx.vnic.resources.ServicePoints().ServicePointHandler(
		msg.ServiceName(), msg.ServiceArea())
	if !ok {
		rx.vnic.resources.Logger().Error("RX: No service point was found for ",
			msg.ServiceName, ":", msg.ServiceArea)
		return
	}
	// If the handler is transactional, it means it is blocking
	// so we don't want to handle it via the workers
	if handler.Transactional() {
		go rx.handleMessage(msg, pb)
		return
	}
	if msg.Action() == common.Reply {
		request := rx.vnic.requests.getRequest(msg.Sequence(), rx.vnic.resources.SysConfig().LocalUuid)
		request.response = pb
		request.cond.Broadcast()
	} else if msg.Action() == common.Notify {
		resp := rx.vnic.resources.ServicePoints().Notify(pb, rx.vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			rx.vnic.resources.Logger().Error(resp.Error())
		}
	} else {
		//Add bool
		resp := rx.vnic.resources.ServicePoints().Handle(pb, msg.Action(), rx.vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			rx.vnic.resources.Logger().Error(resp.Error())
		}
		if msg.Request() {
			err := rx.vnic.Reply(msg, resp)
			if err != nil {
				rx.vnic.resources.Logger().Error(err)
			}
		}
	}
}
