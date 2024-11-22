package state

import (
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"time"
)

const (
	STATE_TOPIC    = "STATE"
	STATE_ENDPOINT = "/" + STATE_TOPIC
)

func CreateStatesFromConfig(config *types.MessagingConfig, isEdge bool) *types2.States {
	edgeState := &types2.EdgeState{}
	if isEdge {
		edgeState.Uuid = config.Local_Uuid
		edgeState.SwitchUuid = config.RemoteUuid
	} else {
		edgeState.Uuid = config.RemoteUuid
		edgeState.SwitchUuid = config.Local_Uuid
	}
	edgeState.UpSince = time.Now().Unix()

	serviceState := &types2.ServiceState{}
	serviceState.Topic = STATE_TOPIC
	serviceState.Edges = make(map[string]bool)
	serviceState.Edges[edgeState.Uuid] = true

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
	m := make(map[string]bool)
	for uuid, exist := range service.Edges {
		m[uuid] = exist
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
