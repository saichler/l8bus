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

func (this *RX) start() {
	go this.readFromSocket()
	go this.notifyRawDataListener()
}

func (this *RX) shutdown() {
	this.shuttingDown = true
	if this.vnic.conn != nil {
		this.vnic.conn.Close()
	}
	this.rx.Shutdown()
}

func (this *RX) name() string {
	return "RX"
}

// loop and Read incoming data from the socket
func (this *RX) readFromSocket() {
	// While the port is active
	for this.vnic.running {
		// read data ([]byte) from socket
		data, err := nets.Read(this.vnic.conn, this.vnic.resources.SysConfig())
		//If therer is an error
		if err != nil {
			if this.vnic.IsVNet {
				break
			}
			if !this.shuttingDown {
				this.vnic.reconnect()
				continue
			} else {
				break
			}
		}
		if data != nil {
			// If still active, write the data to the RX queue
			if this.vnic.running {
				this.rx.Add(data)
			}
		} else {
			// If data is nil, it means the port was shutdown
			// so break and cleanup
			break
		}
	}
	this.vnic.resources.Logger().Debug("RX for ", this.vnic.name, " ended.")
	//Just in case, mark the port as shutdown so other thread will stop as well.
	this.vnic.Shutdown()
}

// Notify the RawDataListener on new data
func (this *RX) notifyRawDataListener() {
	// While the port is active
	for this.vnic.running {
		// Read next data ([]byte) block
		data := this.rx.Next()
		// If data is not nil
		if data != nil {
			this.vnic.stats.RxMsgCount++
			this.vnic.stats.RxDataCont += int64(len(data))
			// if there is a dataListener, this is a switch
			if this.vnic.resources.DataListener() != nil {
				this.vnic.resources.DataListener().HandleData(data, this.vnic)
			} else {
				msg, err := this.vnic.protocol.MessageOf(data)
				if err != nil {
					this.vnic.resources.Logger().Error(err)
					continue
				}
				pb, err := this.vnic.protocol.ElementsOf(msg)
				if err != nil {
					this.vnic.resources.Logger().Error(err)
					if msg.Request() {
						resp := object.NewError(err.Error())
						err = this.vnic.Reply(msg, resp)
						if err != nil {
							this.vnic.resources.Logger().Error(err)
						}
					} else if msg.Reply() {
						resp := object.NewError(err.Error())
						request := this.vnic.requests.getRequest(msg.Sequence(), this.vnic.resources.SysConfig().LocalUuid)
						request.response = resp
						request.cond.Broadcast()
					}
					continue
				}

				//This is a reply message, should not find a handler
				//and just notify
				if msg.Reply() {
					request := this.vnic.requests.getRequest(msg.Sequence(), this.vnic.resources.SysConfig().LocalUuid)
					request.response = pb
					request.cond.Broadcast()
					continue
				}
				// Otherwise call the handler per the action & the type
				// If Reauest == blocking, hence run in a go routing.
				if msg.Request() {
					go this.handleMessage(msg, pb)
				} else {
					this.handleMessage(msg, pb)
				}
			}
		}
	}
	this.vnic.resources.Logger().Debug("ND for ", this.vnic.name, " has Ended")
	this.vnic.Shutdown()
}

func (this *RX) handleMessage(msg common.IMessage, pb common.IElements) {
	if msg.Action() == common.Reply {
		request := this.vnic.requests.getRequest(msg.Sequence(), this.vnic.resources.SysConfig().LocalUuid)
		request.response = pb
		request.cond.Broadcast()
	} else if msg.Action() == common.Notify {
		resp := this.vnic.resources.ServicePoints().Notify(pb, this.vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			panic(this.vnic.resources.SysConfig().LocalAlias + " " + resp.Error().Error())
			this.vnic.resources.Logger().Error(resp.Error())
		}
	} else {
		//Add bool
		resp := this.vnic.resources.ServicePoints().Handle(pb, msg.Action(), this.vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			panic(this.vnic.resources.SysConfig().LocalAlias + " " + resp.Error().Error())
			this.vnic.resources.Logger().Error(resp.Error())
		}
		if msg.Request() {
			err := this.vnic.Reply(msg, resp)
			if err != nil {
				this.vnic.resources.Logger().Error(err)
			}
		}
	}
}
