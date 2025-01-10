package protocol

import (
	"github.com/saichler/serializer/go/serializers"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync/atomic"
)

// Running sequence number for the messages
var sequence atomic.Int32
var mar interfaces.Serializer

func init() {
	mar = serializers.Default
}

func GenerateHeader(msg *types.Message) []byte {
	header := make([]byte, 109)
	for i, c := range msg.SourceUuid {
		header[i] = byte(c)
	}
	for i, c := range msg.SourceSwitchUuid {
		header[i+36] = byte(c)
	}
	for i, c := range msg.Destination {
		header[i+72] = byte(c)
	}
	header[108] = byte(msg.Priority)
	return header
}

func HeaderOf(data []byte) (string, string, string, types.Priority) {
	// Source will always be Uuid
	source := string(data[0:36])
	// Source switch will always be Uuid
	sourceSwitch := string(data[36:72])
	//Destination, either than being a uuid can also be a topic/multicast so it might not be full 16 bytes
	dest := make([]byte, 0)
	for i := 72; i < 108; i++ {
		if data[i] == 0 {
			break
		}
		dest = append(dest, data[i])
	}
	pri := types.Priority(data[108])
	return source, sourceSwitch, string(dest), pri
}

func MessageOf(data []byte, registry interfaces.ITypeRegistry) (*types.Message, error) {
	msg, err := mar.Unmarshal(data[109:], "Message", registry)
	if err != nil {
		panic(err)
	}
	return msg.(*types.Message), err
}

func ProtoOf(msg *types.Message, registry interfaces.ITypeRegistry) (proto.Message, error) {
	data, err := interfaces.SecurityProvider().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	info, err := registry.TypeInfo(msg.Type)
	if err != nil {
		panic(err)
		return nil, interfaces.Error(err)
	}
	pbIns, err := info.NewInstance()
	if err != nil {
		return nil, err
	}

	pb := pbIns.(proto.Message)
	err = proto.Unmarshal(data, pb)
	return pb, err
}

func CreateMessageFor(priority types.Priority, action types.Action, source, sourceSwitch, dest string, pb proto.Message, registry interfaces.ITypeRegistry) ([]byte, error) {
	//first marshal the protobuf into bytes
	data, err := mar.Marshal(pb, nil)
	if err != nil {
		return nil, err
	}
	//Encode the data
	encData, err := interfaces.SecurityProvider().Encrypt(data)
	if err != nil {
		return nil, err
	}
	//create the wrapping message for the destination
	msg := &types.Message{}
	msg.SourceUuid = source
	msg.SourceSwitchUuid = sourceSwitch
	msg.Destination = dest
	msg.Sequence = sequence.Add(1)
	msg.Priority = priority
	msg.Data = encData
	msg.Type = reflect.ValueOf(pb).Elem().Type().Name()
	msg.Action = action
	//Now serialize the message
	msgData, err := mar.Marshal(msg, nil)
	if err != nil {
		return nil, err
	}
	//Create the header for the switch
	header := GenerateHeader(msg)
	//Append the msgData to the header
	header = append(header, msgData...)
	return header, nil
}
