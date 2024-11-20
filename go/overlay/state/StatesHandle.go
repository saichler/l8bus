package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"google.golang.org/protobuf/proto"
)

func (ssp *StatesServicePoint) Post(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	states := pb.(*types.States)
	ssp.MergeState(states)
	return nil, nil
}

func (ssp *StatesServicePoint) Put(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (ssp *StatesServicePoint) Patch(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (ssp *StatesServicePoint) Delete(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (ssp *StatesServicePoint) Get(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (ssp *StatesServicePoint) EndPoint() string {
	return "/States"
}
func (ssp *StatesServicePoint) Topic() string {
	return STATE_TOPIC
}
