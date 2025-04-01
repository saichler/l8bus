package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/serializer/go/serialize/object"
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
	switchTable.desc = "SwitchTable (" + switchService.resources.SysConfig().LocalUuid + ") - "
	return switchTable
}

func (this *SwitchTable) uniCastToAll(serviceName string, serviceArea int32, action types.Action, pb proto.Message) {
	conns := this.conns.all()
	mobjects := object.New(nil, pb)
	data, err := this.switchService.protocol.CreateMessageFor("", serviceName, serviceArea, types.Priority_P1, action,
		this.switchService.resources.SysConfig().LocalUuid,
		this.switchService.resources.SysConfig().LocalUuid, mobjects, false, false, this.switchService.protocol.NextMessageNumber(), nil)
	if err != nil {
		this.switchService.resources.Logger().Error("Failed to create message to send to all: ", err)
		return
	}
	for _, vnic := range conns {
		this.switchService.resources.Logger().Trace(this.desc, "sending message to ", vnic.Resources().SysConfig().RemoteUuid)
		vnic.SendMessage(data)
	}
}

func (this *SwitchTable) addVNic(vnic common.IVirtualNetworkInterface) {
	config := vnic.Resources().SysConfig()
	//check if this port is local to the machine, e.g. not belong to public subnet
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	// If it is local, add it to the internal map
	if isLocal && !config.ForceExternal {
		this.conns.addInternal(config.RemoteUuid, vnic)
	} else {
		// otherwise, add it to the external connections
		this.conns.addExternal(config.RemoteUuid, vnic)
	}

	hc := health.Health(this.switchService.resources)
	hp := hc.HealthPoint(config.RemoteUuid)
	if hp == nil {
		hp = this.newHealthPoint(config)
	} else {
		this.mergeServices(hp, config)
	}
	hc.Add(hp)

	if !(isLocal && !config.ForceExternal) {
		time.Sleep(time.Millisecond * 100)
		allHealthPoints := hc.All()
		for _, hpe := range allHealthPoints {
			vnic.Multicast(health.ServiceName, 0, types.Action_POST, hpe)
		}
	}
}

func (this *SwitchTable) mergeServices(hp *types.HealthPoint, config *types.SysConfig) {
	if hp.Services == nil {
		hp.Services = config.Services
		return
	}
	if hp.Services.ServiceToAreas == nil {
		hp.Services.ServiceToAreas = config.Services.ServiceToAreas
		return
	}
	for k1, v1 := range config.Services.ServiceToAreas {
		exist, ok := hp.Services.ServiceToAreas[k1]
		if !ok {
			hp.Services.ServiceToAreas[k1] = v1
		} else {
			for k2, v2 := range v1.Areas {
				exist.Areas[k2] = v2
			}
		}

	}
}

func (this *SwitchTable) newHealthPoint(config *types.SysConfig) *types.HealthPoint {
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
	if uuidsMap != nil && sourceSwitch != this.switchService.resources.SysConfig().LocalUuid {
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
