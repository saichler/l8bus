package vnic

import (
	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8types/go/types/l8services"
)

func (this *VirtualNetworkInterface) NotifyServiceAdded(serviceNames []string, serviceArea byte) error {
	curr := health.HealthOf(this.resources.SysConfig().LocalUuid, this.resources)

	hp := &l8health.L8Health{}
	hp.AUuid = curr.AUuid
	hp.Services = curr.Services
	mergeServices(hp, this.resources.SysConfig().Services)
	//send notification for health service
	err := this.Unicast(this.resources.SysConfig().RemoteUuid, health.ServiceName, 0, ifs.PATCH, hp)

	return err
}

func (this *VirtualNetworkInterface) NotifyServiceRemoved(serviceName string, serviceArea byte) error {
	curr := health.HealthOf(this.resources.SysConfig().LocalUuid, this.resources)
	hp := &l8health.L8Health{}
	hp.AUuid = curr.AUuid
	hp.Services = curr.Services
	mergeServices(hp, this.resources.SysConfig().Services)
	ifs.RemoveService(hp.Services, serviceName, int32(serviceArea))
	return this.Unicast(this.resources.SysConfig().RemoteUuid, health.ServiceName, serviceArea, ifs.PATCH, hp)
}

func (this *VirtualNetworkInterface) PropertyChangeNotification(set *l8notify.L8NotificationSet) {
	protocol.AddPropertyChangeCalled(set, this.resources.SysConfig().LocalAlias)
	this.Multicast(set.ServiceName, byte(set.ServiceArea), ifs.Notify, set)
}

func mergeServices(hp *l8health.L8Health, services *l8services.L8Services) {
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
