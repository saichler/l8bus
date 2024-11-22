package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	sharedTypes "github.com/saichler/shared/go/types"
	"sync"
)

type StatesServicePoint struct {
	mtx        *sync.RWMutex
	states     *types.States
	localState *types.States
	localUuid  string
}

func NewStatesServicePoint(registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *StatesServicePoint {
	ssp := &StatesServicePoint{}
	ssp.mtx = &sync.RWMutex{}
	ssp.states = &types.States{}
	ssp.states.Edges = make(map[string]*types.EdgeState)
	ssp.states.Services = make(map[string]*types.ServiceState)

	registry.RegisterStruct(&types.States{})
	err := servicePoints.RegisterServicePoint(&types.States{}, ssp, registry)
	if err != nil {
		panic(err)
	}
	return ssp
}

func (ssp *StatesServicePoint) CreateLocalState(config *sharedTypes.MessagingConfig) {
	ssp.mtx.Lock()
	defer ssp.mtx.Unlock()
	ssp.localState = createStatesFromConfig(config, true)
	ssp.localUuid = config.Local_Uuid
}

func (ssp *StatesServicePoint) LocalState() types.States {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	return *ssp.localState
}

func (ssp *StatesServicePoint) States() types.States {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	return *ssp.states
}

func (ssp *StatesServicePoint) RegisterTopic(topic string) {
	ssp.mtx.Lock()
	defer ssp.mtx.Unlock()
	ssp.localState.Services[topic] = &types.ServiceState{}
	ssp.localState.Services[topic].Edges = make(map[string]string)
	ssp.localState.Services[topic].Edges[ssp.localUuid] = ""
}

func (ssp *StatesServicePoint) UpdateTopicsSwitch(switchUuid string) {
	ssp.mtx.Lock()
	defer ssp.mtx.Unlock()
	updatedMap := make(map[string]*types.ServiceState)
	for topic, state := range ssp.localState.Services {
		updatedMap[topic] = &types.ServiceState{}
		updatedMap[topic].Topic = topic
		updatedMap[topic].Edges = make(map[string]string)
		for edgeUuid, _ := range state.Edges {
			updatedMap[topic].Edges[edgeUuid] = switchUuid
		}
	}
	ssp.localState.Services = updatedMap
}
