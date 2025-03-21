package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"google.golang.org/protobuf/proto"
	"time"
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

func (this *SwitchTable) uniCastToAll(serviceName string, serviceArea int32, action types.Action, pb proto.Message) {
	conns := this.conns.all()
	data, err := this.switchService.protocol.CreateMessageFor("", serviceName, serviceArea, types.Priority_P1, action,
		this.switchService.resources.Config().LocalUuid,
		this.switchService.resources.Config().LocalUuid, pb, false, false, this.switchService.protocol.NextMessageNumber(), nil)
	if err != nil {
		this.switchService.resources.Logger().Error("Failed to create message to send to all: ", err)
		return
	}
	for _, vnic := range conns {
		this.switchService.resources.Logger().Trace(this.desc, "sending message to ", vnic.Resources().Config().RemoteUuid)
		vnic.SendMessage(data)
	}
}

func (this *SwitchTable) addVNic(vnic common.IVirtualNetworkInterface) {
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

	hp := this.newHealthPoint(config)

	hc := health.Health(this.switchService.resources)
	hc.Add(hp)

	for _, healthPoint := range hc.All() {
		err := vnic.Multicast(health.ServiceName, 0, types.Action_POST, healthPoint)
		if err != nil {
			this.switchService.resources.Logger().Error(err)
		}
	}
}

func (this *SwitchTable) newHealthPoint(config *types.VNicConfig) *types.HealthPoint {
	hp := &types.HealthPoint{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.ZUuid = config.LocalUuid
	hp.Status = types.HealthState_Up
	hp.Services = config.Services
	hp.StartTime = time.Now().UnixMilli()
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	hp.IsVnet = config.ForceExternal || !isLocal
	return hp
}

func (this *SwitchTable) ServiceUuids(serviceName string, serviceArea int32, sourceSwitch string) map[string]bool {
	h := health.Health(this.switchService.resources)
	uuidsMap := h.Uuids(serviceName, serviceArea, false)
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
