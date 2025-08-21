package vnet

import (
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
)

type SwitchTable struct {
	conns         *Connections
	services      *Services
	routeTable    *RouteTable
	switchService *VNet
	desc          string
}

func newSwitchTable(switchService *VNet) *SwitchTable {
	switchTable := &SwitchTable{}
	vnetUuid := switchService.resources.SysConfig().LocalUuid
	switchTable.routeTable = newRouteTable(vnetUuid)
	switchTable.conns = newConnections(vnetUuid, switchTable.routeTable, switchService.resources.Logger())
	switchTable.services = newServices(switchTable.routeTable)
	switchTable.switchService = switchService
	switchTable.desc = "SwitchTable (" + switchService.resources.SysConfig().LocalUuid + ") - "
	go switchTable.monitor()
	return switchTable
}

func (this *SwitchTable) addVNic(vnic ifs.IVNic) {
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
	hp := hc.Health(config.RemoteUuid)
	if hp == nil {
		hp = this.newHealth(config)
		hc.Add(hp, false)
	} else {
		this.mergeServices(hp, config)
		hc.Update(hp, false)
	}

	this.switchService.publishRoutes()
}

func (this *SwitchTable) mergeServices(hp *types.Health, config *types.SysConfig) {
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

func (this *SwitchTable) newHealth(config *types.SysConfig) *types.Health {
	hp := &types.Health{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.Status = types.HealthState_Up
	hp.Services = config.Services
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	hp.IsVnet = config.ForceExternal || !isLocal

	if !hp.IsVnet {
		hp.StartTime = time.Now().UnixMilli()
		hp.ZUuid = config.LocalUuid
	}
	sd := &types.ServiceData{ServiceName: health.ServiceName, ServiceArea: 0, ServiceUuid: hp.AUuid}
	this.services.addService(sd)
	for k, v := range hp.Services.ServiceToAreas {
		for k2, _ := range v.Areas {
			sd = &types.ServiceData{ServiceName: k, ServiceArea: k2, ServiceUuid: hp.AUuid}
			this.services.addService(sd)
		}
	}

	return hp
}

func (this *SwitchTable) connectionsForService(serviceName string, serviceArea byte, sourceSwitch string, mode ifs.MulticastMode) map[string]ifs.IVNic {
	isHope0 := this.switchService.resources.SysConfig().LocalUuid == sourceSwitch
	result := make(map[string]ifs.IVNic)
	switch mode {
	case ifs.M_All:
		uuidMap := this.services.serviceUuids(serviceName, serviceArea)
		for uuid, _ := range uuidMap {
			usedUuid, vnic := this.conns.getConnection(uuid, isHope0)
			if vnic != nil {
				result[usedUuid] = vnic
			}
		}
		return result
	default:
		uuid := this.services.serviceFor(serviceName, serviceArea, sourceSwitch, mode)
		if uuid != "" {
			usedUuid, vnic := this.conns.getConnection(uuid, isHope0)
			result[usedUuid] = vnic
			return result
		}
	}
	return this.connectionsForService(serviceName, serviceArea, sourceSwitch, ifs.M_All)
}

func (this *SwitchTable) shutdown() {
	conns := this.conns.all()
	for _, conn := range conns {
		conn.Shutdown()
	}
}

func (this *SwitchTable) monitor() {
	if true {
		return
	}
	for this.switchService.running {
		time.Sleep(time.Second * 15)
		hc := health.Health(this.switchService.resources)
		if hc == nil {
			continue
		}
		allDown := this.conns.allDownConnections()
		for uuid, _ := range allDown {
			this.conns.shutdownConnection(uuid)
			hp := hc.Health(uuid)
			if hp.Status != types.HealthState_Down {
				this.switchService.resources.Logger().Info("Update health status to Down")
				hp.Status = types.HealthState_Down
				hc.Update(hp, false)
			}
		}
	}
}
