package vnet

import (
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	resources2 "github.com/saichler/l8utils/go/utils/resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

func (this *VNet) ConnectNetworks(host string, destPort uint32) error {
	sec := this.resources.Security()
	// Dial the destination and validate the secret and key
	conn, err := sec.CanDial(host, destPort)
	if err != nil {
		return err
	}

	config := &types.SysConfig{MaxDataSize: resources2.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize:   resources2.DEFAULT_QUEUE_SIZE,
		TxQueueSize:   resources2.DEFAULT_QUEUE_SIZE,
		VnetPort:      destPort,
		LocalUuid:     this.resources.SysConfig().LocalUuid,
		Services:      this.resources.SysConfig().Services,
		ForceExternal: true,
		LocalAlias:    this.resources.SysConfig().LocalAlias,
	}

	resources := resources2.NewResources(this.resources.Registry(),
		this.resources.Security(),
		this.resources.Services(),
		this.resources.Logger(),
		this,
		this.resources.Serializer(ifs.BINARY),
		config,
		this.resources.Introspector())

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)

	err = sec.ValidateConnection(conn, config)
	if err != nil {
		return err
	}

	vnic.Start()
	this.notifyNewVNic(vnic)
	return nil
}
