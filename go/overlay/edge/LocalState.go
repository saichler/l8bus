package edge

import (
	"github.com/saichler/layer8/go/types"
)

func (edge *EdgeImpl) RegisterTopic(topic string) {
	edge.localState.Services[topic] = &types.ServiceState{}
	edge.localState.Services[topic].Edges = make(map[string]bool)
	edge.localState.Services[topic].Edges[edge.config.Local_Uuid] = true
}

func (edge *EdgeImpl) updateRemoteUuid() {
	for _, state := range edge.localState.Edges {
		state.SwitchUuid = edge.config.RemoteUuid
	}
}
