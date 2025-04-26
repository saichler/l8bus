package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"sync"
	"time"
)

type SwitchTable struct {
	conns         *Connections
	switchService *VNet
	routes        map[string]string
	desc          string

	lastNTime     int64
	lastNMtx      *sync.Mutex
	pendingNotify bool
}

func newSwitchTable(switchService *VNet) *SwitchTable {
	switchTable := &SwitchTable{}
	switchTable.conns = newConnections(switchService.resources.Logger())
	switchTable.switchService = switchService
	switchTable.desc = "SwitchTable (" + switchService.resources.SysConfig().LocalUuid + ") - "
	switchTable.lastNMtx = &sync.Mutex{}
	switchTable.lastNTime = time.Now().UnixMilli()
	return switchTable
}

func (this *SwitchTable) unicastHealthNotification(serviceName string, serviceArea uint16, action common.Action, set *types.NotificationSet, isLocal bool) {
	mobjects := object.New(nil, set)
	nextId := this.switchService.protocol.NextMessageNumber()
	data, err := this.switchService.protocol.CreateMessageFor("", serviceName, serviceArea, common.P1, action,
		this.switchService.resources.SysConfig().LocalUuid,
		this.switchService.resources.SysConfig().LocalUuid, mobjects, false, false, nextId, nil)
	if err != nil {
		this.switchService.resources.Logger().Error("Failed to create message to send to all: ", err)
		return
	}
	var conns map[string]common.IVirtualNetworkInterface
	if isLocal {
		conns = this.conns.all()
	} else {
		conns = this.conns.allInternals()
	}
	for _, vnic := range conns {
		this.switchService.resources.Logger().Trace(this.desc, "sending message ", nextId, " to ",
			vnic.Resources().SysConfig().RemoteUuid)
		vnic.SendMessage(data)
	}
}

func (this *SwitchTable) unicastAll(isLocal, isForceExternal bool, nextId uint32, data []byte) {
	var conns map[string]common.IVirtualNetworkInterface
	if isLocal && !isForceExternal {
		conns = this.conns.all()
	} else {
		conns = this.conns.allInternals()
	}
	for _, vnic := range conns {
		this.switchService.resources.Logger().Trace(this.desc, "Unicast message ", nextId, " to ",
			vnic.Resources().SysConfig().RemoteUuid)
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
	hc.Add(hp, false)
	this.lastNTime = time.Now().UnixMilli()
	go this.notifyNewVNic()
}

func (this *SwitchTable) notifyNewVNic() {
	this.lastNMtx.Lock()
	if this.pendingNotify {
		defer this.lastNMtx.Unlock()
		return
	}
	this.pendingNotify = true
	this.lastNMtx.Unlock()

	for time.Now().UnixMilli()-this.lastNTime < 300 {
		time.Sleep(time.Millisecond * 100)
	}

	this.pendingNotify = false
	hc := health.Health(this.switchService.resources)
	allHealthPoints := hc.All()
	vnetUuid := this.switchService.resources.SysConfig().LocalUuid
	conns := this.conns.all()
	for _, hpe := range allHealthPoints {
		nextId := this.switchService.protocol.NextMessageNumber()
		healthData, _ := this.switchService.protocol.CreateMessageFor("", health.ServiceName, 0, common.P1,
			common.PATCH, vnetUuid, vnetUuid, object.New(nil, hpe), false, false,
			nextId, nil)
		nextId = this.switchService.protocol.NextMessageNumber()
		syncData, _ := this.switchService.protocol.CreateMessageFor("", health.ServiceName, 0, common.P1,
			common.Sync, vnetUuid, vnetUuid, object.New(nil, hpe), false, false,
			nextId, nil)
		for _, vnic := range conns {
			this.switchService.resources.Logger().Trace(this.desc, "Unicast message ", nextId, " to ",
				vnic.Resources().SysConfig().RemoteUuid)
			vnic.SendMessage(healthData)
			go func() {
				time.Sleep(time.Second)
				vnic.SendMessage(syncData)
			}()
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
	hp.Status = types.HealthState_Up
	hp.Services = config.Services
	isLocal := protocol.IpSegment.IsLocal(config.Address)
	hp.IsVnet = config.ForceExternal || !isLocal

	if !hp.IsVnet {
		hp.StartTime = time.Now().UnixMilli()
		hp.ZUuid = config.LocalUuid
	}
	return hp
}

func (this *SwitchTable) ServiceUuids(serviceName string, serviceArea uint16, sourceSwitch string) map[string]bool {
	h := health.Health(this.switchService.resources)
	uuidsMap := h.Uuids(serviceName, serviceArea)
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
