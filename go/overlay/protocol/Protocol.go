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
	msg, err := this.serializer.Unmarshal(data[109:], "Message", this.resources.Registry())
	if err != nil {
		panic(err)
	}
	return msg.(*types.Message), err
}

func (this *Protocol) ProtoOf(msg *types.Message) (proto.Message, error) {
	data, err := this.resources.Security().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	info, err := this.resources.Registry().Info(msg.Type)
	if err != nil {
		return nil, this.resources.Logger().Error(err)
	}
	pbIns, err := info.NewInstance()
	if err != nil {
		return nil, err
	}

	pb := pbIns.(proto.Message)
	err = proto.Unmarshal(data, pb)
	return pb, err
}

func (this *Protocol) CreateMessageFor(area int32, topic string, priority types.Priority,
	action types.Action, source, sourceVnet string, any interface{}) ([]byte, error) {

	//first marshal the protobuf into bytes
	//Expecting a crash here if it is not a protocol buffer
	//Will implement generic serializer via registry in the future
	pb := any.(proto.Message)
	data, err := this.serializer.Marshal(pb, nil)
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
	msg.SourceUuid = source
	msg.SourceVnetUuid = sourceVnet
	msg.Area = area
	msg.Topic = topic
	msg.Sequence = this.sequence.Add(1)
	msg.Priority = priority
	msg.Data = encData
	msg.Type = reflect.ValueOf(pb).Elem().Type().Name()
	msg.Action = action
	return this.DataFromMessage(msg)
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
