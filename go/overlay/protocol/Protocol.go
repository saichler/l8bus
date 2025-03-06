package protocol

import (
	"github.com/saichler/serializer/go/serialize/serializers"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync/atomic"
)

type Protocol struct {
	sequence   atomic.Int32
	resources  interfaces.IResources
	serializer interfaces.ISerializer
}

func New(resources interfaces.IResources) *Protocol {
	p := &Protocol{}
	p.resources = resources
	p.serializer = p.resources.Serializer(interfaces.BINARY)
	if p.serializer == nil {
		p.serializer = &serializers.ProtoBuffBinary{}
	}
	return p
}

func (this *Protocol) Serializer() interfaces.ISerializer {
	return this.serializer
}

func (this *Protocol) MessageOf(data []byte) (*types.Message, error) {
	msg, err := this.serializer.Unmarshal(data[HEADER_SIZE:], "Message", this.resources.Registry())
	return msg.(*types.Message), err
}

func (this *Protocol) ProtoOf(msg *types.Message) (proto.Message, error) {
	return ProtoOf(msg, this.resources)
}

func ProtoOf(msg *types.Message, resourcs interfaces.IResources) (proto.Message, error) {
	data, err := resourcs.Security().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	typ := msg.Type
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

func (this *Protocol) DataFor(any interface{}) (string, error) {
	var data []byte
	var err error
	//first marshal the protobuf into bytes
	pb, ok := any.(proto.Message)
	if ok {
		data, err = this.serializer.Marshal(pb, nil)
		if err != nil {
			return "", err
		}
	} else {
		data = []byte{}
	}
	//Encode the data
	encData, err := this.resources.Security().Encrypt(data)
	if err != nil {
		return "", err
	}
	return encData, err
}

func (this *Protocol) CreateMessageFor(vlan int32, topic string, priority types.Priority,
	action types.Action, source, sourceVnet string, any interface{}, isRequest, isReply bool, msgNum int32, tr *types.Transaction) ([]byte, error) {

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
	msg.SourceUuid = source
	msg.SourceVnetUuid = sourceVnet
	msg.Vlan = vlan
	msg.Topic = topic
	msg.Sequence = msgNum
	msg.Priority = priority
	msg.Data = encData
	if pb == nil {
		msg.Type = topic
	} else {
		msg.Type = reflect.ValueOf(pb).Elem().Type().Name()
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
