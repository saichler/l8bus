package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	sharedTypes "github.com/saichler/shared/go/types"
)

const (
	STATE_TOPIC = "STATE"
)

var emptyServices = make([]string, 0)

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

func (stc *StateCenter) AddEdge(newEdge interfaces.IEdge) {
	config := newEdge.Config()
	ok := stc.statesServicePoint.edgeExist(config.RemoteUuid)
	interfaces.Debug(stc.desc, "adding Edge ", config.RemoteUuid, " ", config.IsAdjacentASwitch)
	if !ok {
		stc.statesServicePoint.addEdgeFromConfig(&config)
	}
}

func (stc *StateCenter) ServiceUuids(destination, source string) []string {
	return stc.statesServicePoint.serviceUuids(destination, source)
}

func (stc *StateCenter) StateRequest() (*sharedTypes.Request, *types.States) {
	request := &sharedTypes.Request{}
	request.Type = sharedTypes.Action_POST
	return request, stc.statesServicePoint.cloneStates()
}
