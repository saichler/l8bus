package protocol

import (
	"github.com/saichler/shared/go/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync/atomic"
)

// Running sequence number for the messages
var sequence atomic.Int32

func GenerateHeader(msg *types.Message) []byte {
	header := make([]byte, 73)
	for i, c := range msg.SourceUuid {
		header[i] = byte(c)
	}
	for i, c := range msg.Destination {
		header[i+36] = byte(c)
	}
	header[72] = byte(msg.Priority)
	return header
}

func HeaderOf(data []byte) (string, string, types.Priority) {
	//Source will always be Uuid
	source := string(data[0:36])
	//Destination, either than being a uuid can also be a topic/multicast so it might not be full 16 bytes
	dest := make([]byte, 0)
	for i := 36; i < 72; i++ {
		if data[i] == 0 {
			break
		}
		dest = append(dest, data[i])
	}
	pri := types.Priority(data[72])
	return source, string(dest), pri
}

func MessageOf(data []byte) (*types.Message, error) {
	msg := &types.Message{}
	err := proto.Unmarshal(data[73:], msg)
	return msg, err
}

func ProtoOf(msg *types.Message, registry interfaces.IRegistry) (proto.Message, error) {
	data, err := interfaces.SecurityProvider().Decrypt(msg.Data)
	if err != nil {
		return nil, err
	}

	pb, err := registry.NewProtobufInstance(msg.Type)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(data, pb)
	return pb, err
}

func CreateMessageFor(priority types.Priority, request *types.Request, source, dest string, pb proto.Message) ([]byte, error) {
	//first marshal the protobuf into bytes
	data, err := proto.Marshal(pb)
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
	msg.Destination = dest
	msg.Sequence = sequence.Add(1)
	msg.Priority = int32(priority)
	msg.Data = encData
	msg.Type = reflect.ValueOf(pb).Elem().Type().Name()
	msg.Request = request
	//Now serialize the message
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	//Create the header for the switch
	header := GenerateHeader(msg)
	//Append the msgData to the header
	header = append(header, msgData...)
	return header, nil
}
