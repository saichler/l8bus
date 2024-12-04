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
	edges         *Edges
	switchService *SwitchService
	routes        map[string]string
	desc          string
}

func newSwitchTable(switchService *SwitchService) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.edges = newEdges()
	switchTable.switchService = switchService
	switchTable.desc = "SwitchTable (" + switchService.switchConfig.Local_Uuid + ") - "
	return switchTable
}

func (switchTable *SwitchTable) sendToAll(topic string, action types.Action, pb proto.Message) {
	edges := switchTable.edges.all()
	data, err := protocol.CreateMessageFor(types.Priority_P0, action, switchTable.switchService.switchConfig.Local_Uuid,
		switchTable.switchService.switchConfig.Local_Uuid, topic, pb,
		switchTable.switchService.registry)
	if err != nil {
		logs.Error("Failed to create message to send to all: ", err)
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
		if switchTable.switchService.switchConfig.Local_Uuid == config.RemoteUuid {
			panic("Remote")
		}
		// otherwise, add it to the external edges
		switchTable.edges.addExternal(config.RemoteUuid, edge)
		logs.Info(switchTable.desc, "added external edge:", config.RemoteUuid)
	}
	switchTable.switchService.statesServicePoint().AddNewSwitchEdge(&config, switchTable.desc)
	states := switchTable.switchService.statesServicePoint().CloneStates()
	go switchTable.sendToAll(state.STATE_TOPIC, types.Action_POST, states)
}

func (switchTable *SwitchTable) ServiceUuids(destination, sourceSwitch string) map[string]bool {
	uuidsMap := switchTable.switchService.statesServicePoint().ServiceUuids(destination)
	if uuidsMap != nil && sourceSwitch != switchTable.switchService.switchConfig.Local_Uuid {
		// When the message source is not within this switch,
		// we should not publish to adjacent as the overlay is o one hope
		// publish.
		switchTable.edges.removeExternals(uuidsMap)
	}
	return uuidsMap
}
