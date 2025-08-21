package vnic

import (
	"github.com/saichler/l8types/go/ifs"
)

func (this *VirtualNetworkInterface) Proximity(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_Proximity, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) ProximityRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request("", serviceName, serviceArea, action, any, ifs.P8, ifs.M_Proximity)
}

func (this *VirtualNetworkInterface) RoundRobin(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_RoundRobin, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) RoundRobinRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request("", serviceName, serviceArea, action, any, ifs.P8, ifs.M_RoundRobin)
}

func (this *VirtualNetworkInterface) Local(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_Local, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) LocalRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request("", serviceName, serviceArea, action, any, ifs.P8, ifs.M_Local)
}

func (this *VirtualNetworkInterface) Leader(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_Leader, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) LeaderRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request("", serviceName, serviceArea, action, any, ifs.P8, ifs.M_Leader)
}
