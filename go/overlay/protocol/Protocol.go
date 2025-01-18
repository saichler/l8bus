package protocol

import (
	"github.com/saichler/serializer/go/serializers"
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

func (this *Protocol) CreateMessageFor(priority types.Priority, action types.Action, source, sourceSwitch, dest string, pb proto.Message) ([]byte, error) {
	//first marshal the protobuf into bytes
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
	msg.SourceSwitchUuid = sourceSwitch
	msg.Destination = dest
	msg.Sequence = this.sequence.Add(1)
	msg.Priority = priority
	msg.Data = encData
	msg.Type = reflect.ValueOf(pb).Elem().Type().Name()
	msg.Action = action
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
