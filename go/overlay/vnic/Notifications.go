package vnic

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
)

func (this *VirtualNetworkInterface) NotifyServiceAdded() error {
	hc := health.Health(this.resources)
	curr := hc.HealthPoint(this.resources.SysConfig().LocalUuid)
	hp := &types.HealthPoint{}
	hp.AUuid = curr.AUuid
	hp.Services = curr.Services
	mergeServices(hp, this.resources.SysConfig().Services)
	return this.Unicast(this.resources.SysConfig().RemoteUuid, health.ServiceName, 0, common.PATCH, hp)
}

func (this *VirtualNetworkInterface) NotifyServiceRemoved(serviceName string, serviceArea uint16) error {
	hc := health.Health(this.resources)
	curr := hc.HealthPoint(this.resources.SysConfig().LocalUuid)
	hp := &types.HealthPoint{}
	hp.AUuid = curr.AUuid
	hp.Services = curr.Services
	mergeServices(hp, this.resources.SysConfig().Services)
	common.RemoveService(hp.Services, serviceName, int32(serviceArea))
	return this.Unicast(this.resources.SysConfig().RemoteUuid, health.ServiceName, 0, common.PATCH, hp)
}

func (this *VirtualNetworkInterface) PropertyChangeNotification(set *types.NotificationSet) {
	protocol.AddPropertyChangeCalled()
	this.Multicast(set.ServiceName, uint16(set.ServiceArea), common.Notify, set)
}

func mergeServices(hp *types.HealthPoint, services *types.Services) {
	if hp.Services == nil {
		hp.Services = services
		return

	}
	for serviceName, serviceAreas := range services.ServiceToAreas {
		_, ok := hp.Services.ServiceToAreas[serviceName]
		if !ok {
			hp.Services.ServiceToAreas[serviceName] = serviceAreas
			continue
		}
		if hp.Services.ServiceToAreas[serviceName].Areas == nil {
			hp.Services.ServiceToAreas[serviceName].Areas = serviceAreas.Areas
			continue
		}
		for svArea, score := range serviceAreas.Areas {
			serviceArea := svArea
			hp.Services.ServiceToAreas[serviceName].Areas[serviceArea] = score
		}
	}
}
