package vnet

import (
	"time"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8sysconfig"
	"github.com/saichler/l8types/go/types/l8system"
	"github.com/saichler/l8utils/go/utils/ipsegment"
)

func (this *VNet) addHealthForVNic(config *l8sysconfig.L8SysConfig) {
	serviceData := &l8system.L8ServiceData{}
	serviceData.ServiceName = health.ServiceName
	serviceData.ServiceArea = int32(health.ServiceAreaByConfig(config))
	serviceData.ServiceUuid = config.RemoteUuid
	this.switchTable.services.addService(serviceData)

	hp := health.HealthOf(config.RemoteUuid, this.resources)
	hs, _ := health.HealthService(this.resources)
	if hp == nil {
		hp = this.newHealth(config)
		hs.Post(object.New(nil, hp), nil)
	} else {
		this.mergeServices(hp, config)
		hs.Patch(object.New(nil, hp), nil)
	}
}

func (this *VNet) newHealth(config *l8sysconfig.L8SysConfig) *l8health.L8Health {
	hp := &l8health.L8Health{}
	hp.Alias = config.RemoteAlias
	hp.AUuid = config.RemoteUuid
	hp.Status = l8health.L8HealthState_Up
	hp.Services = config.Services
	hp.Stats = &l8health.L8HealthStats{}
	hp.Stats.TxMsgCount = -1
	hp.Stats.RxMsgCount = -1
	hp.Stats.LastMsgTime = -1
	hp.Stats.CpuUsage = -1
	hp.Stats.RxDataCont = -1
	hp.Stats.TxDataCount = -1
	hp.Stats.MemoryUsage = 1
	isLocal := ipsegment.IpSegment.IsLocal(config.Address)
	hp.IsVnet = config.ForceExternal || !isLocal

	if !hp.IsVnet {
		hp.StartTime = time.Now().UnixMilli()
		hp.ZUuid = config.LocalUuid
	}

	return hp
}

func (this *VNet) mergeServices(hp *l8health.L8Health, config *l8sysconfig.L8SysConfig) {
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

func (this *VNet) sendHealthReport(uuid string) {
	time.Sleep(time.Second)
	all := health.All(this.resources)
	this.vnic.Unicast(uuid, health.ServiceName, 0, ifs.PATCH, all)
}
