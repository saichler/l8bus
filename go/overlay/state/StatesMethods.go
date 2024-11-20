package state

import (
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"time"
)

func (ssp *StatesServicePoint) edgeExist(uuid string) bool {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	_, ok := ssp.states.Edges[uuid]
	return ok
}

func createStatesFromConfig(config *types.MessagingConfig, isEdge bool) *types2.States {
	edgeState := &types2.EdgeState{}
	if isEdge {
		edgeState.Uuid = config.Local_Uuid
	} else {
		edgeState.Uuid = config.RemoteUuid
	}
	edgeState.UpSince = time.Now().Unix()

	serviceState := &types2.ServiceState{}
	serviceState.Topic = STATE_TOPIC
	serviceState.Edges = make(map[string]string)
	if isEdge {
		serviceState.Edges[edgeState.Uuid] = config.RemoteUuid
	} else {
		serviceState.Edges[edgeState.Uuid] = config.Local_Uuid
	}
	states := &types2.States{}
	states.Edges = make(map[string]*types2.EdgeState)
	states.Services = make(map[string]*types2.ServiceState)
	states.Edges[edgeState.Uuid] = edgeState
	states.Services[STATE_TOPIC] = serviceState
	return states
}

func (ssp *StatesServicePoint) addEdgeFromConfig(config *types.MessagingConfig, isEdge bool) {
	ssp.MergeState(createStatesFromConfig(config, isEdge))
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
				if uuid == zSide {
					panic("Remote and local are the same")
				}
				ssp.states.Services[topic].Edges[uuid] = zSide
			}
		}
	}
}

func (ssp *StatesServicePoint) serviceUuids(destination string) map[string]string {
	ssp.mtx.RLock()
	defer ssp.mtx.RUnlock()
	service, ok := ssp.states.Services[destination]
	if !ok {
		interfaces.Debug("No Services found for destination: ", destination)
		return nil
	}
	return service.Edges
}

func (ssp *StatesServicePoint) cloneStates() *types2.States {
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

func cloneEdge(edge types2.EdgeState) *types2.EdgeState {
	return &edge
}

func cloneService(service types2.ServiceState) *types2.ServiceState {
	m := make(map[string]string)
	for uuid, zSide := range service.Edges {
		m[uuid] = zSide
	}
	service.Edges = m
	return &service
}
