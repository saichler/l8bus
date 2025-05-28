package protocol

import (
	"github.com/saichler/l8types/go/ifs"
	"time"
)

type MessageHeader struct {
	source      [36]byte
	vnet        [36]byte
	destination [36]byte
	serviceArea uint16
	serviceName string
}

type MessageBody struct {
	sequence    uint32
	priority    ifs.Priority
	action      ifs.Action
	timeout     uint16
	request     bool
	reply       bool
	aaaId       string
	failMessage string
	data        string
	tr          *Transaction
}

type Transaction struct {
	id        [36]byte
	state     ifs.TransactionState
	errMsg    string
	startTime int64
}

func NewTransaction() ifs.ITransaction {
	tr := &Transaction{}
	copy(tr.id[0:36], ifs.NewUuid())
	tr.state = ifs.Create
	tr.startTime = time.Now().Unix()
	return tr
}

type Message struct {
	MessageHeader
	MessageBody
}

func (this *Message) Clone() *Message {
	clone := &Message{}
	clone.source = this.source
	clone.vnet = this.vnet
	clone.destination = this.destination
	clone.serviceArea = this.serviceArea
	clone.serviceName = this.serviceName
	clone.sequence = this.sequence
	clone.priority = this.priority
	clone.action = this.action
	clone.reply = this.reply
	clone.request = this.request
	clone.data = this.data
	clone.failMessage = this.failMessage
	clone.timeout = this.timeout
	clone.aaaId = this.aaaId
	if !ifs.IsNil(this.tr) {
		clone.tr = &Transaction{
			id:        this.tr.id,
			state:     this.tr.state,
			errMsg:    this.tr.errMsg,
			startTime: this.tr.startTime,
		}
	}
	return clone
}

func (this *Message) ReplyClone(resources ifs.IResources) ifs.IMessage {
	reply := this.Clone()
	reply.action = ifs.Reply
	copy(reply.destination[0:36], this.source[0:36])
	copy(reply.source[0:36], resources.SysConfig().LocalUuid)
	copy(reply.vnet[0:36], resources.SysConfig().RemoteUuid)
	reply.request = false
	reply.reply = true
	return reply
}

func (this *Message) FailClone(failMessage string) ifs.IMessage {
	fail := this.Clone()
	fail.failMessage = failMessage
	copy(fail.source[0:36], this.destination[0:36])
	copy(fail.destination[0:36], this.source[0:36])
	return fail
}
