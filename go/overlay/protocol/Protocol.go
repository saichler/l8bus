package protocol

import (
	"encoding/base64"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"sync/atomic"
)

type Protocol struct {
	sequence  atomic.Uint32
	resources ifs.IResources
}

func New(resources ifs.IResources) *Protocol {
	p := &Protocol{}
	p.resources = resources
	object.MessageSerializer = &MessageSerializer{}
	object.TransactionSerializer = &TransactionSerializer{}
	return p
}

func (this *Protocol) MessageOf(data []byte, resources ifs.IResources) (ifs.IMessage, error) {
	msg, _ := object.MessageSerializer.Unmarshal(data, resources)
	return msg.(ifs.IMessage), nil
}

func (this *Protocol) ElementsOf(msg ifs.IMessage) (ifs.IElements, error) {
	return ElementsOf(msg, this.resources)
}

func ElementsOf(msg ifs.IMessage, resourcs ifs.IResources) (ifs.IElements, error) {

	data, err := base64.StdEncoding.DecodeString(msg.Data())
	if err != nil {
		return nil, err
	}
	result := &object.Elements{}
	err = result.Deserialize(data, resourcs.Registry())
	if err != nil {
		return nil, err
	}
	return result, err
}

func (this *Protocol) NextMessageNumber() uint32 {
	return this.sequence.Add(1)
}

func DataFor(elems ifs.IElements, security ifs.ISecurityProvider) (string, error) {
	var data []byte
	var err error

	data, err = elems.Serialize()
	return base64.StdEncoding.EncodeToString(data), err
}

func (this *Protocol) CreateMessageFor(destination, serviceName string, serviceArea uint16,
	priority ifs.Priority, action ifs.Action, source, vnet string, o ifs.IElements,
	isRequest, isReply bool, msgNum uint32, tr ifs.ITransaction) ([]byte, error) {

	AddMessageCreated()

	var data []byte
	var err error

	data, err = o.Serialize()
	if err != nil {
		return nil, err
	}

	//create the wrapping message for the destination
	msg := &Message{}
	copy(msg.source[0:36], source)
	copy(msg.vnet[0:36], vnet)
	copy(msg.destination[0:36], destination)
	msg.serviceName = serviceName
	msg.serviceArea = serviceArea
	msg.sequence = msgNum
	msg.priority = priority
	msg.data = base64.StdEncoding.EncodeToString(data)
	msg.action = action
	msg.request = isRequest
	msg.reply = isReply
	msg.tr, _ = tr.(*Transaction)
	return object.MessageSerializer.Marshal(msg, this.resources)
}

func (this *Protocol) CreateMessageForm(msg ifs.IMessage, o ifs.IElements) ([]byte, error) {
	var data []byte
	var err error

	data, err = o.Serialize()
	if err != nil {
		return nil, err
	}

	//create the wrapping message for the destination
	msg.SetData(base64.StdEncoding.EncodeToString(data))
	return object.MessageSerializer.Marshal(msg, this.resources)
}
