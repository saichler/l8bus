package vnic

import "github.com/saichler/l8types/go/ifs"

func (this *VirtualNetworkInterface) Multicast(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_All, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) multicast(priority ifs.Priority, multicastMode ifs.MulticastMode, serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	elems, err := createElements(any, this.resources)
	if err != nil {
		return err
	}
	return this.components.TX().Multicast("", serviceName, serviceArea, action, elems, priority, multicastMode,
		false, false, this.protocol.NextMessageNumber(), ifs.Empty, "", "", -1, "")
}
