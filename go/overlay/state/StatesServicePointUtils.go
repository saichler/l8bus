package state

import (
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"time"
)

const (
	STATE_TOPIC = "STATE"
)

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

func Print(states *types2.States, uuid string) {
	interfaces.Info("Review ", uuid)
	for _, edge := range states.Edges {
		interfaces.Info("  ", edge.Uuid)
	}
	for topic, service := range states.Services {
		interfaces.Info("  ", topic)
		for uuid, _ := range service.Edges {
			interfaces.Info("      ", uuid)
		}
	}
}
