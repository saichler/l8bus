package state

import (
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
)

func (ssp *StatesServicePoint) edgeExist(uuid string) bool {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	_, ok := ssp.states.Edges[uuid]
	return ok
}

func (ssp *StatesServicePoint) addEdgeFromConfig(config *types.MessagingConfig, isEdge bool) {
	ssp.MergeState(CreateStatesFromConfig(config, isEdge))
}

func (ssp *StatesServicePoint) MergeState(states *types2.States) {
	ssp.mtx.Lock()
	defer ssp.mtx.Unlock()

	for uuid, edgeState := range states.Edges {
		existEdgeState, ok := ssp.states.Edges[uuid]
		if !ok {
			ssp.states.Edges[uuid] = edgeState
		} else {
			existEdgeState.UpSince = edgeState.UpSince
			existEdgeState.LastMessage = edgeState.LastMessage
			existEdgeState.MessagesReceived = edgeState.MessagesReceived
			existEdgeState.MessagesSent = edgeState.MessagesSent
		}
	}

	for topic, serviceState := range states.Services {
		_, ok := ssp.states.Services[topic]
		if !ok {
			ssp.states.Services[topic] = serviceState
		} else {
			for uuid, zSide := range serviceState.Edges {
				ssp.states.Services[topic].Edges[uuid] = zSide
			}
		}
	}
}

func (ssp *StatesServicePoint) ServiceUuids(destination string) map[string]bool {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	service, ok := ssp.states.Services[destination]
	if !ok {
		interfaces.Debug("No Services found for destination: ", destination)
		return nil
	}
	return service.Edges
}

func (ssp *StatesServicePoint) CloneStates() *types2.States {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	clone := &types2.States{}
	clone.Edges = ssp.cloneEdgeStateMap()
	clone.Services = ssp.cloneServiceStateMap()
	return clone
}

func (ssp *StatesServicePoint) cloneEdgeStateMap() map[string]*types2.EdgeState {
	edges := make(map[string]*types2.EdgeState)
	for uuid, state := range ssp.states.Edges {
		edges[uuid] = cloneEdge(*state)
	}
	return edges
}

func (ssp *StatesServicePoint) cloneServiceStateMap() map[string]*types2.ServiceState {
	services := make(map[string]*types2.ServiceState)
	for topic, serviceState := range ssp.states.Services {
		services[topic] = cloneService(*serviceState)
	}
	return services
}

func (ssp *StatesServicePoint) AddNewSwitchEdge(config *types.MessagingConfig, switchTableName string) {
	ok := ssp.edgeExist(config.RemoteUuid)
	interfaces.Debug(switchTableName, "adding Edge ", config.RemoteUuid, " ", config.IsAdjacentASwitch)
	if !ok {
		ssp.addEdgeFromConfig(config, false)
	}
}

func (ssp *StatesServicePoint) FindSwitch(edgeUuid string) string {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	edge, ok := ssp.states.Edges[edgeUuid]
	if ok {
		return edge.SwitchUuid
	}
	panic(edgeUuid)
	return ""
}
