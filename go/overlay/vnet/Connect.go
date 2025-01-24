package vnet

import (
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	resources2 "github.com/saichler/shared/go/share/resources"
)

func (this *VNet) ConnectNetworks(host string, destPort uint32) error {
	sec := this.resources.Security()
	// Dial the destination and validate the secret and key
	conn, err := sec.CanDial(host, destPort)
	if err != nil {
		return err
	}

	resources := resources2.NewResources(this.resources.Registry(),
		this.resources.Security(),
		this.resources.ServicePoints(),
		this.resources.Logger(),
		this.resources.Config().LocalAlias)
	resources.SetDataListener(this)

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)

	config := resources.Config()
	config.SwitchPort = destPort
	config.LocalUuid = this.resources.Config().LocalUuid
	config.Topics = resources.ServicePoints().Topics()
	config.ForceExternal = true

	err = sec.ValidateConnection(conn, config)
	if err != nil {
		return err
	}

	vnic.Start()

	this.notifyNewVNic(vnic)
	return nil
}
