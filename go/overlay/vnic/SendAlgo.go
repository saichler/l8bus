package vnic

import (
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/layer8/go/overlay/health"
)

type SendAlgo int

const (
	Proximity  SendAlgo = 1
	RoundRobin SendAlgo = 2
	Local      SendAlgo = 3
	Leader     SendAlgo = 4
)

func (this *VirtualNetworkInterface) Proximity(serviceName string, serviceArea byte, action ifs.Action, any interface{}) (string, error) {
	return this.send(serviceName, serviceArea, action, any, Proximity)
}

func (this *VirtualNetworkInterface) ProximityRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request(serviceName, serviceArea, action, any, Proximity)
}

func (this *VirtualNetworkInterface) RoundRobin(serviceName string, serviceArea byte, action ifs.Action, any interface{}) (string, error) {
	return this.send(serviceName, serviceArea, action, any, RoundRobin)
}

func (this *VirtualNetworkInterface) RoundRobinRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request(serviceName, serviceArea, action, any, RoundRobin)
}

func (this *VirtualNetworkInterface) Local(serviceName string, serviceArea byte, action ifs.Action, any interface{}) (string, error) {
	return this.send(serviceName, serviceArea, action, any, Local)
}

func (this *VirtualNetworkInterface) LocalRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request(serviceName, serviceArea, action, any, Local)
}

func (this *VirtualNetworkInterface) Leader(serviceName string, serviceArea byte, action ifs.Action, any interface{}) (string, error) {
	return this.send(serviceName, serviceArea, action, any, Leader)
}

func (this *VirtualNetworkInterface) LeaderRequest(serviceName string, serviceArea byte, action ifs.Action, any interface{}) ifs.IElements {
	return this.request(serviceName, serviceArea, action, any, Leader)
}

func (this *VirtualNetworkInterface) send(serviceName string, serviceArea byte, action ifs.Action, any interface{}, algo SendAlgo) (string, error) {
	hc := health.Health(this.resources)
	destination := ""
	switch algo {
	case Proximity:
		destination = hc.ProximityFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case RoundRobin:
		destination = hc.RoundRobinFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case Local:
		destination = hc.LocalFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case Leader:
		destination = hc.LeaderFor(serviceName, serviceArea)
	}

	if destination == "" {
		destination = ifs.DESTINATION_Single
	}

	hp := hc.Health(destination)
	alias := "Unknown Yet"
	if hp != nil {
		alias = hp.Alias
	}
	this.Resources().Logger().Info("Sending Proximity to ", destination, " alias ", alias)

	return destination, this.Unicast(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) request(serviceName string, serviceArea byte, action ifs.Action, any interface{}, algo SendAlgo) ifs.IElements {
	hc := health.Health(this.resources)
	destination := ""
	switch algo {
	case Proximity:
		destination = hc.ProximityFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case RoundRobin:
		destination = hc.RoundRobinFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case Local:
		destination = hc.LocalFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid)
	case Leader:
		destination = hc.LeaderFor(serviceName, serviceArea)
	}
	if destination == "" {
		destination = ifs.DESTINATION_Single
	}

	hp := hc.Health(destination)
	alias := "Unknown Yet"
	if hp != nil {
		alias = hp.Alias
	}
	this.Resources().Logger().Info("Sending Proximity Request to ", destination, " alias ", alias)
	return this.Request(destination, serviceName, serviceArea, action, any)
}
