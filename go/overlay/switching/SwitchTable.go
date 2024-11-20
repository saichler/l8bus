package switching

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/state"
	"github.com/saichler/shared/go/share/interfaces"
	logs "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

type SwitchTable struct {
	internalEdges *protocol.EdgeMap
	externalEdges *protocol.EdgeMap
	stateCenter   *state.StateCenter
	switchUuid    string
	desc          string
}

func newSwitchTable(switchUuid string, registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.internalEdges = protocol.NewEdgeMap()
	switchTable.externalEdges = protocol.NewEdgeMap()
	switchTable.stateCenter = state.NewStateCenter(switchUuid, registry, servicePoints)
	switchTable.switchUuid = switchUuid
	switchTable.desc = "SwitchTable (" + switchUuid + ") - "
	return switchTable
}

func (switchTable *SwitchTable) allEdgeList() []interfaces.IEdge {
	edges := make([]interfaces.IEdge, 0)
	switchTable.internalEdges.Iterate(func(k, v interface{}) {
		edge := v.(interfaces.IEdge)
		logs.Trace(switchTable.desc, "collected Internal Edge ", edge.Config().RemoteUuid)
		edges = append(edges, edge)
	})
	switchTable.externalEdges.Iterate(func(k, v interface{}) {
		edge := v.(interfaces.IEdge)
		logs.Trace(switchTable.desc, "collected External Edge ", edge.Config().RemoteUuid)
		edges = append(edges, edge)
	})
	return edges
}

func (switchTable *SwitchTable) broadcast(topic string, action types.Action, pb proto.Message) {
	edges := switchTable.allEdgeList()
	logs.Debug(switchTable.desc, "broadcasting to ", len(edges))

	data, err := protocol.CreateMessageFor(types.Priority_P0, action, switchTable.switchUuid, switchTable.switchUuid, topic, pb)
	if err != nil {
		logs.Error("Failed to send broadcast:", err)
		return
	}
	for _, edge := range edges {
		logs.Trace(switchTable.desc, "sending message to ", edge.Config().RemoteUuid)
		edge.Send(data)
	}
}

func (switchTable *SwitchTable) addEdge(edge interfaces.IEdge) {
	config := edge.Config()
	//check if this port is local to the machine, e.g. not belong to public subnet
	isLocal := ipSegment.isLocal(config.Address)
	// If it is local, add it to the internal map
	if isLocal && !config.IsAdjacentASwitch {
		//check if the port already exist
		ep, ok := switchTable.internalEdges.Get(config.RemoteUuid)
		if ok {
			//If it exists, then shutdown the existing instance as we want the new one to be used.
			ep.Shutdown()
		}
		switchTable.internalEdges.Put(config.RemoteUuid, edge)
		logs.Info(switchTable.desc, "added internal edge:", config.RemoteUuid)
	} else {
		// If it is public, add it to the external map
		// but first check if it already exists
		ep, ok := switchTable.externalEdges.Get(config.RemoteUuid)
		if ok {
			//if it already exists, shut it down.
			ep.Shutdown()
		}
		switchTable.externalEdges.Put(config.RemoteUuid, edge)
		logs.Info(switchTable.desc, "added external edge:", config.RemoteUuid)
	}
	switchTable.stateCenter.AddEdge(edge)
	states := switchTable.stateCenter.States()
	go switchTable.broadcast(state.STATE_TOPIC, types.Action_POST, states)
}

func (switchTable *SwitchTable) fetchEdgeByUuid(id string) interfaces.IEdge {
	p, ok := switchTable.internalEdges.Get(id)
	if !ok {
		p, ok = switchTable.externalEdges.Get(id)
	}
	return p
}

func (switchTable *SwitchTable) ServiceUuids(destination, sourceSwitch string) map[string]string {
	uuidMap := switchTable.stateCenter.ServiceUuids(destination)
	if uuidMap != nil && sourceSwitch != switchTable.switchUuid {
		excludeExternal := make(map[string]string)
		for uuid, remote := range uuidMap {
			_, ok := switchTable.externalEdges.Get(uuid)
			if !ok {
				excludeExternal[uuid] = remote
			}
		}
		return excludeExternal
	}
	return uuidMap
}
