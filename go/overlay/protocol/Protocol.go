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
	return p
}

func (this *Protocol) MessageOf(data []byte, resources ifs.IResources) (*ifs.Message, error) {
	msg := &ifs.Message{}
	_, err := msg.Unmarshal(data, this.resources)
	return msg, err
}

func (this *Protocol) ElementsOf(msg *ifs.Message) (ifs.IElements, error) {
	return ElementsOf(msg, this.resources)
}

func ElementsOf(msg *ifs.Message, resourcs ifs.IResources) (ifs.IElements, error) {

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

func (this *Protocol) CreateMessageFor(destination, serviceName string, serviceArea byte,
	priority ifs.Priority, action ifs.Action, source, vnet string, o ifs.IElements,
	isRequest, isReply bool, msgNum uint32,
	tr_state ifs.TransactionState, tr_id, tr_errMsg string, tr_start int64,
	token string) ([]byte, error) {

	AddMessageCreated()

	var data []byte
	var err error

	data, err = o.Serialize()
	if err != nil {
		return nil, err
	}

	msg, err := this.resources.Security().Message(token)
	if err != nil {
		return nil, err
	}
	msg.Init(destination,
		serviceName,
		serviceArea,
		priority,
		action,
		source,
		vnet,
		base64.StdEncoding.EncodeToString(data),
		isRequest,
		isReply,
		msgNum,
		tr_state,
		tr_id,
		tr_errMsg,
		tr_start)

	return msg.Marshal(nil, this.resources)
}

func (this *Protocol) CreateMessageForm(msg *ifs.Message, o ifs.IElements) ([]byte, error) {
	var data []byte
	var err error

	data, err = o.Serialize()
	if err != nil {
		return nil, err
	}

	//create the wrapping message for the destination
	msg.SetData(base64.StdEncoding.EncodeToString(data))
	return msg.Marshal(nil, this.resources)
}
