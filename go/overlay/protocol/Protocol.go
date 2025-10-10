package protocol

import (
	"sync/atomic"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
)

var Discovery_Enabled = true

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
	result := &object.Elements{}
	err := result.Deserialize(msg.Data(), resourcs.Registry())
	if err != nil {
		return nil, err
	}
	return result, err
}

func (this *Protocol) NextMessageNumber() uint32 {
	return this.sequence.Add(1)
}

func DataFor(elems ifs.IElements, security ifs.ISecurityProvider) ([]byte, error) {
	var data []byte
	var err error

	data, err = elems.Serialize()
	return data, err
}

func (this *Protocol) CreateMessageFor(destination, serviceName string, serviceArea byte,
	priority ifs.Priority, multicastMode ifs.MulticastMode, action ifs.Action, source, vnet string, o ifs.IElements,
	isRequest, isReply bool, msgNum uint32,
	tr_state ifs.TransactionState, tr_id, tr_errMsg string,
	tr_created, tr_queued, tr_running, tr_complete, tr_timeout int64, tr_replica byte, tr_isReplica bool,
	aaaid string) ([]byte, error) {

	AddMessageCreated()

	var data []byte
	var err error

	data, err = o.Serialize()
	if err != nil {
		return nil, err
	}

	msg, err := this.resources.Security().Message(aaaid)
	if err != nil {
		return nil, err
	}
	msg.Init(destination,
		serviceName,
		serviceArea,
		priority,
		multicastMode,
		action,
		source,
		vnet,
		data,
		isRequest,
		isReply,
		msgNum,
		tr_state,
		tr_id,
		tr_errMsg,
		tr_created,
		tr_queued,
		tr_running,
		tr_complete,
		tr_timeout,
		tr_replica,
		tr_isReplica)

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
	msg.SetData(data)
	return msg.Marshal(nil, this.resources)
}
