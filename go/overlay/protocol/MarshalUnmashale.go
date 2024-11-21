package protocol

import (
	"google.golang.org/protobuf/proto"
	"sync"
)

var protoMarshalSync = &sync.Mutex{}

func Marshal(pb proto.Message) ([]byte, error) {
	protoMarshalSync.Lock()
	defer protoMarshalSync.Unlock()
	return proto.Marshal(pb)
}

func Unmarshal(data []byte, pb proto.Message) error {
	protoMarshalSync.Lock()
	defer protoMarshalSync.Unlock()
	return proto.Unmarshal(data, pb)
}
