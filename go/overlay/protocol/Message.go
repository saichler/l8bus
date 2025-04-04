package protocol

import (
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/nets"
	"time"
)

type MessageHeader struct {
	source      [36]byte
	vnet        [36]byte
	destination [36]byte
	serviceArea uint16
	serviceName string
}

type Transaction struct {
	id        [36]byte
	state     common.TransactionState
	errMsg    string
	startTime int64
}

func NewTransaction() common.ITransaction {
	tr := &Transaction{}
	copy(tr.id[0:36], common.NewUuid())
	tr.state = common.Create
	tr.startTime = time.Now().Unix()
	return tr
}

type Message struct {
	MessageHeader
	sequence    uint32
	priority    common.Priority
	action      common.Action
	timeout     uint16
	request     bool
	reply       bool
	failMessage string
	data        string
	tr          *Transaction
}

const (
	POS_Source       = 0
	POS_Vnet         = 36
	POS_Destination  = POS_Vnet + 36
	POS_Service_Area = POS_Destination + 36
	POS_Service_Name = POS_Service_Area + 2
)

func (this *Message) Serialize() []byte {
	POS_Sequence := POS_Service_Name + 2 + len(this.serviceName)
	POS_Priority := POS_Sequence + 4
	POS_Action := POS_Priority + 1
	POS_Timeout := POS_Action + 1
	POS_Request_Reply := POS_Timeout + 2
	POS_Fail_Message := POS_Request_Reply + 1
	POS_DATA := POS_Fail_Message + 2 + len(this.failMessage)
	POS_Tr := POS_DATA + 4 + len(this.data)

	POS_Tr_Id := POS_Tr + 1
	POS_Tr_State := POS_Tr_Id + 36
	POS_Tr_Start_Time := POS_Tr_State + 1
	POS_Tr_Err_Message := POS_Tr_Start_Time + 8
	POS_END := POS_Tr_Id
	if this.tr != nil {
		POS_END = POS_Tr_Err_Message + 2 + len(this.tr.errMsg)
	}

	var data []byte
	if this.tr == nil {
		data = make([]byte, POS_Tr+1)
	} else {
		data = make([]byte, POS_END)
	}

	copy(data[POS_Source:POS_Vnet], this.source[0:36])
	copy(data[POS_Vnet:POS_Destination], this.vnet[0:36])
	copy(data[POS_Destination:POS_Service_Area], this.destination[0:36])
	copy(data[POS_Service_Area:POS_Service_Name], nets.UInt162Bytes(this.serviceArea))
	copy(data[POS_Service_Name:POS_Service_Name+2], nets.UInt162Bytes(uint16(len(this.serviceName))))
	copy(data[POS_Service_Name+2:POS_Sequence], this.serviceName)
	copy(data[POS_Sequence:POS_Priority], nets.UInt322Bytes(this.sequence))
	data[POS_Priority] = byte(this.priority)
	data[POS_Action] = byte(this.action)
	copy(data[POS_Timeout:POS_Request_Reply], nets.UInt162Bytes(this.timeout))
	data[POS_Request_Reply] = nets.ByteOf(this.request, this.reply)
	copy(data[POS_Fail_Message:POS_Fail_Message+2], nets.UInt162Bytes(uint16(len(this.failMessage))))
	copy(data[POS_Fail_Message+2:POS_DATA], this.failMessage)
	copy(data[POS_DATA:POS_DATA+4], nets.UInt322Bytes(uint32(len(this.data))))
	copy(data[POS_DATA+4:POS_Tr], this.data)
	if this.tr == nil {
		data[POS_Tr] = 0
		return data
	}
	data[POS_Tr] = 1
	copy(data[POS_Tr_Id:POS_Tr_State], this.tr.id[0:36])
	data[POS_Tr_State] = byte(this.tr.state)
	copy(data[POS_Tr_Start_Time:POS_Tr_Err_Message], nets.Long2Bytes(this.tr.startTime))
	copy(data[POS_Tr_Err_Message:POS_Tr_Err_Message+2], nets.UInt162Bytes(uint16(len(this.tr.errMsg))))
	copy(data[POS_Tr_Err_Message+2:POS_END], this.tr.errMsg)

	return data
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
	if this.tr != nil {
		clone.tr = &Transaction{
			id:        this.tr.id,
			state:     this.tr.state,
			errMsg:    this.tr.errMsg,
			startTime: this.tr.startTime,
		}
	}
	return clone
}

func (this *Message) ReplyClone(resources common.IResources) common.IMessage {
	reply := this.Clone()
	reply.action = common.Reply
	reply.destination = this.source
	copy(reply.source[0:36], resources.SysConfig().LocalUuid)
	copy(reply.vnet[0:36], resources.SysConfig().RemoteUuid)
	reply.request = false
	reply.reply = true
	return reply
}

func (this *Message) FailClone(failMessage string) common.IMessage {
	fail := this.Clone()
	fail.failMessage = failMessage
	fail.source = this.destination
	fail.destination = this.source
	return fail
}

func HeaderOf(data []byte) (string, string, string, string, uint16, common.Priority) {

	size := nets.Bytes2UInt16(data[POS_Service_Name : POS_Service_Name+2])
	POS_Sequence := POS_Service_Name + 2 + int(size)
	POS_Priority := POS_Sequence + 4

	return string(data[POS_Source:POS_Vnet]),
		string(data[POS_Vnet:POS_Destination]),
		string(data[POS_Destination:POS_Service_Area]),
		string(data[POS_Service_Name+2 : POS_Sequence]),
		nets.Bytes2UInt16(data[POS_Service_Area:POS_Service_Name]),
		common.Priority(data[POS_Priority])
}

func Deserialize(data []byte) *Message {
	msg := &Message{}
	copy(msg.source[0:36], data[POS_Source:POS_Vnet])
	copy(msg.vnet[0:36], data[POS_Vnet:POS_Destination])
	copy(msg.destination[0:36], data[POS_Destination:POS_Service_Area])
	msg.serviceArea = nets.Bytes2UInt16(data[POS_Service_Area:POS_Service_Name])
	size := nets.Bytes2UInt16(data[POS_Service_Name : POS_Service_Name+2])
	POS_Sequence := POS_Service_Name + 2 + int(size)
	POS_Priority := POS_Sequence + 4
	POS_Action := POS_Priority + 1
	POS_Timeout := POS_Action + 1
	POS_Request_Reply := POS_Timeout + 2
	msg.serviceName = string(data[POS_Service_Name+2 : POS_Sequence])

	msg.sequence = nets.Bytes2UInt32(data[POS_Sequence:POS_Priority])
	msg.priority = common.Priority(data[POS_Priority])
	msg.action = common.Action(data[POS_Action])
	msg.timeout = nets.Bytes2UInt16(data[POS_Timeout:POS_Request_Reply])
	msg.request, msg.reply = nets.BoolOf(data[POS_Request_Reply])

	POS_Fail_Message := POS_Request_Reply + 1
	size = nets.Bytes2UInt16(data[POS_Fail_Message : POS_Fail_Message+2])
	POS_DATA := POS_Fail_Message + 2 + int(size)
	msg.failMessage = string(data[POS_Fail_Message+2 : POS_DATA])

	size = nets.Bytes2UInt16(data[POS_DATA : POS_DATA+2])
	POS_Tr := POS_DATA + 4 + int(size)
	msg.data = string(data[POS_DATA+4 : POS_Tr])
	if data[POS_Tr] == 0 {
		return msg
	}

	POS_Tr_Id := POS_Tr + 1
	POS_Tr_State := POS_Tr_Id + 36
	POS_Tr_Start_Time := POS_Tr_State + 1
	POS_Tr_Err_Message := POS_Tr_Start_Time + 8
	POS_END := POS_Tr_Id

	msg.tr = &Transaction{}
	copy(msg.tr.id[0:36], data[POS_Tr_Id:POS_Tr_State])
	msg.tr.state = common.TransactionState(data[POS_Tr_State])
	msg.tr.startTime = nets.Bytes2Long(data[POS_Tr_Start_Time:POS_Tr_Err_Message])
	size = nets.Bytes2UInt16(data[POS_Tr_Err_Message : POS_Tr_Err_Message+2])
	POS_END = POS_Tr_Err_Message + 2 + int(size)
	msg.tr.errMsg = string(data[POS_Tr_Err_Message+2 : POS_END])

	return msg
}
