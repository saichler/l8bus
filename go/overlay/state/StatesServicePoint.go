package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"google.golang.org/protobuf/proto"
)

type StatesServicePoint struct {
	states *types.States
}

func NewStatesServicePoint(registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *StatesServicePoint {
	ssp := &StatesServicePoint{}
	ssp.states = &types.States{}
	ssp.states.Edges = make(map[string]*types.EdgeState)
	ssp.states.Services = make(map[string]*types.Services)
	registry.RegisterStruct(&types.States{})
	err := servicePoints.RegisterServicePoint(&types.States{}, ssp, registry)
	if err != nil {
		panic(err)
	}
	return ssp
}

func (ssp *StatesServicePoint) Post(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	states := pb.(*types.States)
	for _, edge := range states.Edges {
		interfaces.Logger().Info("Post:", edge.Uuid)
	}
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
	return "/EdgeInfos"
}
