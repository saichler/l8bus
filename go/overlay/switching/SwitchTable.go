package switching

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

type SwitchTable struct {
	conns         *Connections
	switchService *SwitchService
	routes        map[string]string
	desc          string
}

func newSwitchTable(switchService *SwitchService) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.conns = newConnections(switchService.resources.Logger())
	switchTable.switchService = switchService
	switchTable.desc = "SwitchTable (" + switchService.resources.Config().Local_Uuid + ") - "
	return switchTable
}

func (switchTable *SwitchTable) sendToAll(topic string, action types.Action, pb proto.Message) {
	conns := switchTable.conns.all()
	data, err := switchTable.switchService.protocol.CreateMessageFor(types.Priority_P0, action, switchTable.switchService.resources.Config().Local_Uuid,
		switchTable.switchService.resources.Config().Local_Uuid, topic, pb)
	if err != nil {
		switchTable.switchService.resources.Logger().Error("Failed to create message to send to all: ", err)
		return
	}
	for _, vnic := range conns {
		switchTable.switchService.resources.Logger().Trace(switchTable.desc, "sending message to ", vnic.Resources().Config().RemoteUuid)
		vnic.Send(data)
	}
}

func (switchTable *SwitchTable) addVNic(vnic interfaces.IVirtualNetworkInterface) {
	config := vnic.Resources().Config()
	//check if this port is local to the machine, e.g. not belong to public subnet
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	// If it is local, add it to the internal map
	if isLocal && !config.ForceExternal {
		switchTable.conns.addInternal(config.RemoteUuid, vnic)
	} else {
		// otherwise, add it to the external connections
		switchTable.conns.addExternal(config.RemoteUuid, vnic)
	}

	hp := &types2.HealthPoint{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.ZUuid = config.Local_Uuid
	hp.Services = vnic.Resources().Config().Topics

	hc := health.Health(switchTable.switchService.resources)
	hc.Add(hp)
	for _, p := range hc.AllPoints() {
		vnic.Do(types.Action_POST, health.TOPIC, p)
	}
	switchTable.sendToAll(health.TOPIC, types.Action_POST, hp)
}

func (switchTable *SwitchTable) ServiceUuids(destination, sourceSwitch string) map[string]bool {
	h := health.Health(switchTable.switchService.resources)
	uuidsMap := h.UuidsForTopic(destination)
	if uuidsMap != nil && sourceSwitch != switchTable.switchService.resources.Config().Local_Uuid {
		// When the message source is not within this switch,
		// we should not publish to adjacent as the overlay is o one hope
		// publish.
		switchTable.conns.filterExternals(uuidsMap)
	}
	return uuidsMap
}

func (switchTable *SwitchTable) shutdown() {
	conns := switchTable.conns.all()
	for _, conn := range conns {
		conn.Shutdown()
	}
}
