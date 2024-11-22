package edge

import (
	"github.com/saichler/layer8/go/types"
)

func (edge *EdgeImpl) RegisterTopic(topic string) {
	edge.localState.Services[topic] = &types.ServiceState{}
	edge.localState.Services[topic].Edges = make(map[string]string)
	edge.localState.Services[topic].Edges[edge.config.Local_Uuid] = ""
}

func (edge *EdgeImpl) updateRemoteUuid() {
	updatedMap := make(map[string]*types.ServiceState)
	for topic, state := range edge.localState.Services {
		updatedMap[topic] = &types.ServiceState{}
		updatedMap[topic].Topic = topic
		updatedMap[topic].Edges = make(map[string]string)
		for edgeUuid, _ := range state.Edges {
			updatedMap[topic].Edges[edgeUuid] = edge.config.RemoteUuid
		}
	}
	edge.localState.Services = updatedMap
}
