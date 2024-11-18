package state

import (
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/types"
	"time"
)

func (ssp *StatesServicePoint) edgeExist(uuid string) bool {
	_, ok := ssp.states.Edges[uuid]
	return ok
}

func (ssp *StatesServicePoint) addEdge(config *types.MessagingConfig) {
	edgeState := &types2.EdgeState{}
	edgeState.Uuid = config.RemoteUuid
	edgeState.UpSince = time.Now().Unix()

	ssp.states.Edges[config.RemoteUuid] = edgeState

	services := &types2.Services{}
	services.Uuids = make(map[string]bool)
	services.Uuids[config.RemoteUuid] = true

	edgeState.Services = make(map[string]*types2.Services)
	edgeState.Services[config.RemoteUuid] = services

	stateService, ok := ssp.states.Services[STATE_TOPIC]
	if !ok {
		stateService = &types2.Services{}
		stateService.Uuids = make(map[string]bool)
		ssp.states.Services[STATE_TOPIC] = stateService
	}
	stateService.Uuids[config.RemoteUuid] = true
}

func (ssp *StatesServicePoint) serviceUuids(destination, source string) []string {
	service, ok := ssp.states.Services[destination]
	if !ok {
		return emptyServices
	}
	result := make([]string, len(service.Uuids))
	i := 0
	for uuid, _ := range service.Uuids {
		if uuid != source {
			result[i] = uuid
		}
		i++
	}
	return result
}

func (ssp *StatesServicePoint) cloneStates() *types2.States {
	clone := &types2.States{}
	clone.Edges = ssp.cloneEdges()
	clone.Services = cloneServicesMap(ssp.states.Services)
	return clone
}

func (ssp *StatesServicePoint) cloneEdges() map[string]*types2.EdgeState {
	edges := make(map[string]*types2.EdgeState)
	for uuid, state := range ssp.states.Edges {
		edges[uuid] = cloneEdge(*state)
	}
	return edges
}

func cloneEdge(edge types2.EdgeState) *types2.EdgeState {
	edge.Services = cloneServicesMap(edge.Services)
	return &edge
}

func cloneServicesMap(servicesMap map[string]*types2.Services) map[string]*types2.Services {
	services := make(map[string]*types2.Services)
	for topic, svcs := range servicesMap {
		services[topic] = cloneServices(svcs)
	}
	return services
}

func cloneServices(svcs *types2.Services) *types2.Services {
	services := &types2.Services{}
	services.Uuids = make(map[string]bool)
	for uuid, exist := range svcs.Uuids {
		services.Uuids[uuid] = exist
	}
	return services
}
