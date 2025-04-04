package protocol

import (
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/nets"
)

var MSer = &MessageSerializer{}
var TSer = &TransactionSerializer{}

type MessageSerializer struct {
}

func (this *MessageSerializer) Mode() common.SerializerMode {
	return common.BINARY
}

func (this *MessageSerializer) Marshal(any interface{}, r common.IRegistry) ([]byte, error) {
	message := any.(*Message)
	POS_Sequence := POS_Service_Name + 2 + len(message.serviceName)
	POS_Priority := POS_Sequence + 4
	POS_Action := POS_Priority + 1
	POS_Timeout := POS_Action + 1
	POS_Request_Reply := POS_Timeout + 2
	POS_Fail_Message := POS_Request_Reply + 1
	POS_DATA := POS_Fail_Message + 2 + len(message.failMessage)
	POS_Tr := POS_DATA + 4 + len(message.data)

	trData, _ := TSer.Marshal(message.Tr(), nil)

	data := make([]byte, POS_Tr+len(trData))

	copy(data[POS_Source:POS_Vnet], message.source[0:36])
	copy(data[POS_Vnet:POS_Destination], message.vnet[0:36])
	destSize := len(message.destination)
	data[POS_Destination] = byte(destSize)
	if destSize > 0 {
		copy(data[POS_Destination+1:POS_Service_Area], message.destination[0:36])
	}
	copy(data[POS_Service_Area:POS_Service_Name], nets.UInt162Bytes(message.serviceArea))
	copy(data[POS_Service_Name:POS_Service_Name+2], nets.UInt162Bytes(uint16(len(message.serviceName))))
	copy(data[POS_Service_Name+2:POS_Sequence], message.serviceName)
	copy(data[POS_Sequence:POS_Priority], nets.UInt322Bytes(message.sequence))
	data[POS_Priority] = byte(message.priority)
	data[POS_Action] = byte(message.action)
	copy(data[POS_Timeout:POS_Request_Reply], nets.UInt162Bytes(message.timeout))
	data[POS_Request_Reply] = nets.ByteOf(message.request, message.reply)
	copy(data[POS_Fail_Message:POS_Fail_Message+2], nets.UInt162Bytes(uint16(len(message.failMessage))))
	copy(data[POS_Fail_Message+2:POS_DATA], message.failMessage)
	copy(data[POS_DATA:POS_DATA+4], nets.UInt322Bytes(uint32(len(message.data))))
	copy(data[POS_DATA+4:POS_Tr], message.data)

	copy(data[POS_Tr:], trData)

	return data, nil
}

func (this *MessageSerializer) Unmarshal(data []byte, r common.IRegistry) (interface{}, error) {
	msg := &Message{}
	copy(msg.source[0:36], data[POS_Source:POS_Vnet])
	copy(msg.vnet[0:36], data[POS_Vnet:POS_Destination])
	destSize := data[POS_Destination]
	if destSize > 0 {
		msg.destination = string(data[POS_Destination+1 : POS_Service_Area])
	}
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

	size32 := nets.Bytes2UInt32(data[POS_DATA : POS_DATA+4])
	POS_Tr := POS_DATA + 4 + int(size32)
	msg.data = string(data[POS_DATA+4 : POS_Tr])
	if data[POS_Tr] == 0 {
		return msg, nil
	}

	tr, _ := TSer.Unmarshal(data[POS_Tr:], nil)
	msg.tr = tr.(*Transaction)

	return msg, nil
}
