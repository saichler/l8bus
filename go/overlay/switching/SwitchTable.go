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
	edges       *Edges
	stateCenter *state.StateCenter
	switchUuid  string
	desc        string
}

func newSwitchTable(switchUuid string, registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.edges = newEdges()
	switchTable.stateCenter = state.NewStateCenter(switchUuid, registry, servicePoints)
	switchTable.switchUuid = switchUuid
	switchTable.desc = "SwitchTable (" + switchUuid + ") - "
	return switchTable
}

func (switchTable *SwitchTable) broadcast(topic string, action types.Action, pb proto.Message) {
	edges := switchTable.edges.all()
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
		switchTable.edges.addInternal(config.RemoteUuid, edge)
		logs.Info(switchTable.desc, "added internal edge:", config.RemoteUuid)
	} else {
		// otherwise, add it to the external edges
		switchTable.edges.addExternal(config.RemoteUuid, edge)
		logs.Info(switchTable.desc, "added external edge:", config.RemoteUuid)
	}
	switchTable.stateCenter.AddEdge(edge)
	states := switchTable.stateCenter.States()
	go switchTable.broadcast(state.STATE_TOPIC, types.Action_POST, states)
}

func (switchTable *SwitchTable) ServiceUuids(destination, sourceSwitch string) map[string]string {
	uuidsMap := switchTable.stateCenter.ServiceUuids(destination)
	if uuidsMap != nil && sourceSwitch != switchTable.switchUuid {
		// When the message source is not within this switch,
		// we should not publish to adjacent as the overlay is o one hope
		// publish.
		switchTable.edges.removeExternals(uuidsMap)
	}
	return uuidsMap
}
