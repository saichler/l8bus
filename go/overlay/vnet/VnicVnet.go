package vnet

import (
	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8types/go/types/l8services"
)

type VnicVnet struct {
	vnet *VNet
}

func newVnicVnet(vnet *VNet) *VnicVnet {
	return &VnicVnet{vnet: vnet}
}

func (this *VnicVnet) Start() {
	panic("implement me")
}

func (this *VnicVnet) Shutdown() {
	panic("implement me")
}

func (this *VnicVnet) Name() string {
	panic("implement me")
	return ""
}

func (this *VnicVnet) SendMessage(data []byte) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Unicast(destination string, serviceName string, serviceArea byte, action ifs.Action, data interface{}) error {
	_, conn := this.vnet.switchTable.conns.getConnection(destination, true)
	elems := object.New(nil, data)
	bts, err := this.vnet.protocol.CreateMessageFor(destination, serviceName, serviceArea, ifs.P1, ifs.M_All, action,
		this.Resources().SysConfig().LocalUuid, this.Resources().SysConfig().LocalUuid, elems,
		false, false, this.vnet.protocol.NextMessageNumber(), ifs.NotATransaction,
		"", "", -1, -1, -1, -1, -1, 0, false, "")
	if err != nil {
		return err
	}
	conn.SendMessage(bts)
	return nil
}

func (this *VnicVnet) Request(destination string, serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Reply(msg *ifs.Message, elements ifs.IElements) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Multicast(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	var err error
	var data []byte
	myUuid := this.vnet.resources.SysConfig().LocalUuid
	connections := this.vnet.switchTable.connectionsForService(serviceName, serviceArea, myUuid, ifs.M_All)
	for uuid, connection := range connections {
		data, err = this.vnet.protocol.CreateMessageFor(uuid, serviceName, serviceArea, ifs.P1, ifs.M_All, action,
			myUuid, uuid, object.New(nil, any), false, false, this.vnet.protocol.NextMessageNumber(),
			ifs.NotATransaction, "", "", -1, -1, -1, -1,
			-1, 0, false, "")
		if err != nil {
			continue
		}
		e := connection.SendMessage(data)
		if e != nil {
			err = e
		}
	}
	return err
}

func (this *VnicVnet) RoundRobin(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) RoundRobinRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Proximity(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) ProximityRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Leader(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) LeaderRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Local(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) LocalRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Forward(msg *ifs.Message, destination string) ifs.IElements {
	panic("implement me")
	return nil
}

func (this *VnicVnet) ServiceAPI(serviceName string, area byte) ifs.ServiceAPI {
	panic("implement me")
	return nil
}

func (this *VnicVnet) Resources() ifs.IResources {
	return this.vnet.resources
}

func (this *VnicVnet) NotifyServiceAdded(serviceNames []string, serviceArea byte) error {
	curr := health.HealthOf(this.vnet.resources.SysConfig().LocalUuid, this.vnet.resources)
	hp := &l8health.L8Health{}
	hp.AUuid = curr.AUuid
	hp.Services = curr.Services
	mergeServices(hp, this.vnet.resources.SysConfig().Services)
	for _, serviceName := range serviceNames {
		this.Multicast(serviceName, serviceArea, ifs.PATCH, hp)
	}
	return nil
}

func (this *VnicVnet) NotifyServiceRemoved(serviceName string, area byte) error {
	panic("implement me")
	return nil
}

func (this *VnicVnet) PropertyChangeNotification(set *l8notify.L8NotificationSet) {
	this.vnet.PropertyChangeNotification(set)
}

func (this *VnicVnet) WaitForConnection() {
	panic("implement me")
}

func (this *VnicVnet) Running() bool {
	panic("implement me")
	return false
}

func (this *VnicVnet) RegisterServiceLink(link *l8services.L8ServiceLink) {
	panic("implement me")
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
