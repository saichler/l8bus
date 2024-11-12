package switching

import (
	"github.com/saichler/overlayK8s/go/protocol"
	"github.com/saichler/overlayK8s/go/state"
	"github.com/saichler/shared/go/interfaces"
	logs "github.com/saichler/shared/go/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

type SwitchTable struct {
	internalEdges *protocol.EdgeMap
	externalEdges *protocol.EdgeMap
	stateCenter   *state.StateCenter
}

func newSwitchTable() *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.internalEdges = protocol.NewEdgeMap()
	switchTable.externalEdges = protocol.NewEdgeMap()
	switchTable.stateCenter = &state.StateCenter{}
	return switchTable
}

func (switchTable *SwitchTable) allEdgeList() []interfaces.IEdge {
	edges := make([]interfaces.IEdge, 0)
	switchTable.internalEdges.Iterate(func(k, v interface{}) {
		edges = append(edges, v.(interfaces.IEdge))
	})
	switchTable.externalEdges.Iterate(func(k, v interface{}) {
		edges = append(edges, v.(interfaces.IEdge))
	})
	return edges
}

func (switchTable *SwitchTable) broadcast(topic string, request *types.Request, switchUuid string, pb proto.Message) {
	logs.Debug("Broadcast")
	edges := switchTable.allEdgeList()
	data, err := protocol.CreateMessageFor(types.Priority_P0, request, switchUuid, topic, pb)
	if err != nil {
		logs.Error("Failed to send broadcast:", err)
		return
	}
	for _, edge := range edges {
		edge.Send(data)
	}
}

func (switchTable *SwitchTable) addEdge(edge interfaces.IEdge, switchUuid string) {
	//check if this port is local to the machine, e.g. not belong to public subnet
	isLocal := ipSegment.isLocal(edge.Addr())
	// If it is local, add it to the internal map
	if isLocal {
		//check if the port already exist
		ep, ok := switchTable.internalEdges.Get(edge.Uuid())
		if ok {
			//If it exists, then shutdown the existing instance as we want the new one to be used.
			ep.Shutdown()
		}
		switchTable.internalEdges.Put(edge.Uuid(), edge)
	} else {
		// If it is public, add it to the external map
		// but first check if it already exists
		ep, ok := switchTable.externalEdges.Get(edge.Uuid())
		if ok {
			//if it already exists, shut it down.
			ep.Shutdown()
		}
		switchTable.externalEdges.Put(edge.Uuid(), edge)
	}
	switchTable.stateCenter.AddEdge(edge)
	request, infos := switchTable.stateCenter.InfosRequest()
	go switchTable.broadcast(state.STATE_TOPIC, request, switchUuid, infos)
}

func (switchTable *SwitchTable) fetchEdgeByUuid(id string) interfaces.IEdge {
	p, ok := switchTable.internalEdges.Get(id)
	if !ok {
		p, ok = switchTable.externalEdges.Get(id)
	}
	return p
}
