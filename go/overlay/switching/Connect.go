package switching

import (
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	resources2 "github.com/saichler/shared/go/share/resources"
)

func (switchService *SwitchService) Switch2Switch(host string, destPort uint32) error {
	sec := switchService.resources.Security()
	// Dial the destination and validate the secret and key
	conn, err := sec.CanDial(host, destPort)
	if err != nil {
		return err
	}

	resources := resources2.NewResources(switchService.resources.Registry(),
		switchService.resources.Security(),
		switchService.resources.ServicePoints(),
		switchService.resources.Logger(),
		switchService.resources.Config().LocalAlias)
	resources.SetDataListener(switchService)
	
	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)

	config := resources.Config()
	config.SwitchPort = destPort
	config.Local_Uuid = switchService.resources.Config().Local_Uuid
	config.Topics = resources.ServicePoints().Topics()
	config.ForceExternal = true

	err = sec.ValidateConnection(conn, config)
	if err != nil {
		return err
	}

	vnic.Start()

	switchService.notifyNewVNic(vnic)
	return nil
}
