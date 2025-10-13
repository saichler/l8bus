package vnet

import (
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8types/go/types/l8services"
)

type VnicVnet struct {
	vnet *VNet
}

func newVnicVnet(vnet *VNet) *VnicVnet {
	return &VnicVnet{vnet: vnet}
}

func (v *VnicVnet) Start() {
	panic("implement me")
}

func (v *VnicVnet) Shutdown() {
	panic("implement me")
}

func (v *VnicVnet) Name() string {
	panic("implement me")
	return ""
}

func (v *VnicVnet) SendMessage(data []byte) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Unicast(destination string, serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Request(destination string, serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Reply(msg *ifs.Message, elements ifs.IElements) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Multicast(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) RoundRobin(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) RoundRobinRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Proximity(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) ProximityRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Leader(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) LeaderRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Local(serviceName string, area byte, action ifs.Action, data interface{}) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) LocalRequest(serviceName string, area byte, action ifs.Action, data interface{}, timeout int, returnAttributes ...string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Forward(msg *ifs.Message, destination string) ifs.IElements {
	panic("implement me")
	return nil
}

func (v *VnicVnet) ServiceAPI(serviceName string, area byte) ifs.ServiceAPI {
	panic("implement me")
	return nil
}

func (v *VnicVnet) Resources() ifs.IResources {
	panic("implement me")
	return nil
}

func (v *VnicVnet) NotifyServiceAdded(serviceNames []string, area byte) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) NotifyServiceRemoved(serviceName string, area byte) error {
	panic("implement me")
	return nil
}

func (v *VnicVnet) PropertyChangeNotification(set *l8notify.L8NotificationSet) {
	panic("implement me")
}

func (v *VnicVnet) WaitForConnection() {
	panic("implement me")
}

func (v *VnicVnet) Running() bool {
	panic("implement me")
	return false
}

func (v *VnicVnet) RegisterServiceLink(link *l8services.L8ServiceLink) {
	panic("implement me")
}
