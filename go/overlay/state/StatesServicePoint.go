package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"sync"
)

type StatesServicePoint struct {
	mtx    *sync.RWMutex
	states *types.States
}

func NewStatesServicePoint(registry interfaces.ITypeRegistry, servicePoints interfaces.IServicePoints) *StatesServicePoint {
	ssp := &StatesServicePoint{}
	ssp.mtx = &sync.RWMutex{}
	ssp.states = &types.States{}
	ssp.states.Edges = make(map[string]*types.EdgeState)
	ssp.states.Services = make(map[string]*types.ServiceState)

	registry.Register(&types.States{})
	err := servicePoints.RegisterServicePoint(&types.States{}, ssp, registry)
	if err != nil {
		panic(err)
	}
	return ssp
}

func (ssp *StatesServicePoint) States() types.States {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	return *ssp.states
}
