package protocol

import (
	"github.com/saichler/serializer/go/serialize/serializers"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync/atomic"
)

type Protocol struct {
	sequence   atomic.Int32
	resources  common.IResources
	serializer common.ISerializer
}

func New(resources common.IResources) *Protocol {
	p := &Protocol{}
	p.resources = resources
	p.serializer = p.resources.Serializer(common.BINARY)
	if p.serializer == nil {
		p.serializer = &serializers.ProtoBuffBinary{}
	}
	return p
}

func (this *Protocol) Serializer() common.ISerializer {
	return this.serializer
}

func (this *Protocol) MessageOf(data []byte) (*types.Message, error) {
	msg, err := this.serializer.Unmarshal(data[HEADER_SIZE:], "Message", this.resources.Registry())
	return msg.(*types.Message), err
}

func (this *Protocol) ProtoOf(msg *types.Message) (proto.Message, error) {
	return ProtoOf(msg, this.resources)
}

func ProtoOf(msg *types.Message, resourcs common.IResources) (proto.Message, error) {
	data, err := resourcs.Security().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	typ := msg.ProtoType
	if msg.Tr != nil && msg.IsReply {
		typ = reflect.TypeOf(types.Transaction{}).Name()
	}

	info, err := resourcs.Registry().Info(typ)
	if err != nil {
		return nil, resourcs.Logger().Error(err)
	}
	pbIns, err := info.NewInstance()
	if err != nil {
		return nil, err
	}

	pb := pbIns.(proto.Message)
	err = proto.Unmarshal(data, pb)

	return pb, err
}

func (this *Protocol) NextMessageNumber() int32 {
	return this.sequence.Add(1)
}

func DataFor(any interface{}, serializer common.ISerializer, security common.ISecurityProvider) (string, error) {
	var data []byte
	var err error
	//first marshal the protobuf into bytes
	pb, ok := any.(proto.Message)
	if ok {
		data, err = serializer.Marshal(pb, nil)
		if err != nil {
			return "", err
		}
	} else {
		data = []byte{}
	}
	//Encode the data
	encData, err := security.Encrypt(data)
	if err != nil {
		return "", err
	}
	return encData, err
}

func (this *Protocol) CreateMessageFor(destination, serviceName string, serviceArea int32,
	priority types.Priority, action types.Action, source, vnet string, any interface{},
	isRequest, isReply bool, msgNum int32, tr *types.Transaction) ([]byte, error) {

	var data []byte
	var err error

	//first marshal the protobuf into bytes
	pb, ok := any.(proto.Message)
	if ok {
		data, err = this.serializer.Marshal(pb, nil)
		if err != nil {
			return nil, err
		}
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
	if pb != nil {
		msg.ProtoType = reflect.ValueOf(pb).Elem().Type().Name()
	}
	msg.Action = action
	msg.IsRequest = isRequest
	msg.IsReply = isReply
	msg.Tr = tr
	d, e := this.DataFromMessage(msg)
	return d, e
}

func (this *Protocol) CreateMessageForm(msg *types.Message, any interface{}) ([]byte, error) {
	//first marshal the protobuf into bytes
	//Expecting a crash here if it is not a protocol buffer
	//Will implement generic serializer via registry in the future
	var data []byte
	var err error
	pb, ok := any.(proto.Message)
	if ok {
		data, err = this.serializer.Marshal(pb, nil)
	}
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
	msgData, err := this.serializer.Marshal(msg, nil)
	if err != nil {
		return nil, err
	}
	//Create the header for the switch
	header := CreateHeader(msg)
	//Append the msgData to the header
	header = append(header, msgData...)
	return header, nil
}
