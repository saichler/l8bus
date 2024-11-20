package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
)

const (
	STATE_TOPIC = "STATE"
)

type StateCenter struct {
	statesServicePoint *StatesServicePoint
	desc               string
}

func NewStateCenter(uuid string, registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *StateCenter {
	stc := &StateCenter{}
	stc.statesServicePoint = NewStatesServicePoint(registry, servicePoints)
	stc.desc = "StateCenter (" + uuid + ") - "
	return stc
}

func (stc *StateCenter) ServicePoint() *StatesServicePoint {
	return stc.statesServicePoint
}

func (stc *StateCenter) AddEdge(newEdge interfaces.IEdge) {
	config := newEdge.Config()
	ok := stc.statesServicePoint.edgeExist(config.RemoteUuid)
	interfaces.Debug(stc.desc, "adding Edge ", config.RemoteUuid, " ", config.IsAdjacentASwitch)
	if !ok {
		stc.statesServicePoint.addEdgeFromConfig(&config, false)
	}
}

func (stc *StateCenter) ServiceUuids(destination string) map[string]string {
	return stc.statesServicePoint.serviceUuids(destination)
}

func (stc *StateCenter) States() *types.States {
	return stc.statesServicePoint.cloneStates()
}
