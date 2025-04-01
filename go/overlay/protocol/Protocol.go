package protocol

import (
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
)

type Protocol struct {
	sequence  atomic.Int32
	resources common.IResources
}

func New(resources common.IResources) *Protocol {
	p := &Protocol{}
	p.resources = resources
	return p
}

func (this *Protocol) MessageOf(data []byte) (*types.Message, error) {
	msg := &types.Message{}
	err := proto.Unmarshal(data[HEADER_SIZE:], msg)
	if err != nil {
		return nil, err
	}
	return msg, err
}

func (this *Protocol) ElementsOf(msg *types.Message) (common.IElements, error) {
	return ElementsOf(msg, this.resources)
}

func ElementsOf(msg *types.Message, resourcs common.IResources) (common.IElements, error) {
	data, err := resourcs.Security().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	mobjects := &types.Elements{}
	err = proto.Unmarshal(data, mobjects)
	if err != nil {
		return nil, err
	}

	result := &object.MObjects{}
	err = result.Deserialize(mobjects, resourcs.Registry())
	if err != nil {
		return nil, err
	}
	return result, err
}

func (this *Protocol) NextMessageNumber() int32 {
	return this.sequence.Add(1)
}

func DataFor(elems common.IElements, security common.ISecurityProvider) (string, error) {
	var data []byte
	var err error

	objs, err := elems.Serialize()
	if err != nil {
		return "", err
	}

	data, err = proto.Marshal(objs)
	if err != nil {
		return "", err
	}

	//Encode the data
	encData, err := security.Encrypt(data)
	if err != nil {
		return "", err
	}
	return encData, err
}

func (this *Protocol) CreateMessageFor(destination, serviceName string, serviceArea int32,
	priority types.Priority, action types.Action, source, vnet string, o common.IElements,
	isRequest, isReply bool, msgNum int32, tr *types.Transaction) ([]byte, error) {

	var data []byte
	var err error

	objs, err := o.Serialize()
	if err != nil {
		return nil, err
	}

	data, err = proto.Marshal(objs)
	if err != nil {
		return nil, err
	}

	//Encode the data
	encData, err := this.resources.Security().Encrypt(data)
	if err != nil {
		return nil, err
	}
	//create the wrapping message for the destination
	msg := &types.Message{}
	msg.Source = source
	msg.SourceVnet = vnet
	msg.Destination = destination
	msg.ServiceName = serviceName
	msg.ServiceArea = serviceArea
	msg.Sequence = msgNum
	msg.Priority = priority
	msg.Data = encData
	msg.Action = action
	msg.IsRequest = isRequest
	msg.IsReply = isReply
	msg.Tr = tr
	d, e := this.DataFromMessage(msg)
	return d, e
}

func (this *Protocol) CreateMessageForm(msg *types.Message, o common.IElements) ([]byte, error) {
	var data []byte
	var err error

	mobjects, err := o.Serialize()
	if err != nil {
		return nil, err
	}

	data, err = proto.Marshal(mobjects)
	if err != nil {
		return nil, err
	}

	//Encode the data
	encData, err := this.resources.Security().Encrypt(data)
	if err != nil {
		return nil, err
	}
	//create the wrapping message for the destination
	msg.Data = encData
	d, e := this.DataFromMessage(msg)
	return d, e
}

func (this *Protocol) DataFromMessage(msg *types.Message) ([]byte, error) {
	//Now serialize the message
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	//Create the header for the switch
	header := CreateHeader(msg)
	//Append the msgData to the header
	header = append(header, msgData...)
	return header, nil
}
