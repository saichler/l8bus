package vnet

import (
	"time"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8sysconfig"
	"github.com/saichler/l8types/go/types/l8system"
	"github.com/saichler/l8utils/go/utils/strings"
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
	switchTable.desc = strings.New("SwitchTable (", switchService.resources.SysConfig().LocalUuid, ") - ").String()
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

	hp := health.HealthOf(config.RemoteUuid, this.switchService.resources)
	hs, _ := health.HealthService(this.switchService.resources)
	if hp == nil {
		hp = this.newHealth(config)
		hs.Put(object.New(nil, hp), this.switchService.vnic)
	} else {
		this.mergeServices(hp, config)
		hs.Patch(object.New(nil, hp), this.switchService.vnic)
	}

	this.switchService.publishRoutes()
}

func (this *SwitchTable) mergeServices(hp *l8health.L8Health, config *l8sysconfig.L8SysConfig) {
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

func (this *SwitchTable) newHealth(config *l8sysconfig.L8SysConfig) *l8health.L8Health {
	hp := &l8health.L8Health{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.Status = l8health.L8HealthState_Up
	hp.Services = config.Services
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	hp.IsVnet = config.ForceExternal || !isLocal

	if !hp.IsVnet {
		hp.StartTime = time.Now().UnixMilli()
		hp.ZUuid = config.LocalUuid
	}

	for k, v := range hp.Services.ServiceToAreas {
		for k2, _ := range v.Areas {
			sd := &l8system.L8ServiceData{ServiceName: k, ServiceArea: k2, ServiceUuid: hp.AUuid}
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
			usedUuid, vnic := this.conns.getConnection(uuid, true)
			if vnic != nil {
				result[usedUuid] = vnic
			} else {
				this.switchService.resources.Logger().Error("Cannot find vnic for uuid:", uuid, ":", usedUuid)
			}
			return result
		} else {
			this.switchService.resources.Logger().Error("Cannot find uuid for service ", serviceName, ":", serviceArea)
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

		allDown := this.conns.allDownConnections()
		for uuid, _ := range allDown {
			this.conns.shutdownConnection(uuid)
			hp := health.HealthOf(uuid, this.switchService.resources)
			if hp.Status != l8health.L8HealthState_Down {
				this.switchService.resources.Logger().Info("Update health status to Down")
				hp.Status = l8health.L8HealthState_Down
				hs, _ := health.HealthService(this.switchService.resources)
				hs.Patch(object.New(nil, hp), this.switchService.vnic)
			}
		}
	}
}
