package vnet

import (
	"fmt"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

type SwitchTable struct {
	conns         *Connections
	switchService *VNet
	routes        map[string]string
	desc          string
}

func newSwitchTable(switchService *VNet) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.conns = newConnections(switchService.resources.Logger())
	switchTable.switchService = switchService
	switchTable.desc = "SwitchTable (" + switchService.resources.Config().LocalUuid + ") - "
	return switchTable
}

func (this *SwitchTable) uniCastToAll(area int32, topic string, action types.Action, pb proto.Message) {
	conns := this.conns.all()
	data, err := this.switchService.protocol.CreateMessageFor(area, topic, types.Priority_P0, action,
		this.switchService.resources.Config().LocalUuid,
		this.switchService.resources.Config().LocalUuid, pb, false, false, this.switchService.protocol.NextMessageNumber())
	if err != nil {
		this.switchService.resources.Logger().Error("Failed to create message to send to all: ", err)
		return
	}
	for _, vnic := range conns {
		this.switchService.resources.Logger().Trace(this.desc, "sending message to ", vnic.Resources().Config().RemoteUuid)
		vnic.SendMessage(data)
	}
}

func (this *SwitchTable) addVNic(vnic interfaces.IVirtualNetworkInterface) {
	config := vnic.Resources().Config()
	//check if this port is local to the machine, e.g. not belong to public subnet
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	// If it is local, add it to the internal map
	if isLocal && !config.ForceExternal {
		this.conns.addInternal(config.RemoteUuid, vnic)
	} else {
		// otherwise, add it to the external connections
		this.conns.addExternal(config.RemoteUuid, vnic)
	}

	hp := &types.HealthPoint{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.ZUuid = config.LocalUuid
	hp.Status = types.HealthState_Up
	hp.ServiceAreas = vnic.Resources().Config().ServiceAreas
	hc := health.Health(this.switchService.resources)
	hc.Add(hp)

	for _, healthPoint := range hc.All() {
		vnic.Multicast(types.CastMode_All, types.Action_POST, 0, health.TOPIC, healthPoint)
	}
	//switchTable.sendToAll(health.TOPIC, types.Action_POST, hp)
}

func (this *SwitchTable) ServiceUuids(area int32, destination, sourceSwitch string) map[string]bool {
	h := health.Health(this.switchService.resources)
	uuidsMap := h.UuidsForTopic(area, destination)
	fmt.Println("Topic:", destination, "UUIds:", uuidsMap)
	if uuidsMap != nil && sourceSwitch != this.switchService.resources.Config().LocalUuid {
		// When the message source is not within this switch,
		// we should not publish to adjacent as the overlay is o one hope
		// publish.
		this.conns.filterExternals(uuidsMap)
	}
	return uuidsMap
}

func (this *SwitchTable) shutdown() {
	conns := this.conns.all()
	for _, conn := range conns {
		conn.Shutdown()
	}
}
