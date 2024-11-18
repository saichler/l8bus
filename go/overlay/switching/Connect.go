package switching

import (
	"github.com/saichler/overlayK8s/go/overlay/edge"
	"github.com/saichler/shared/go/share/interfaces"
)

func (switchService *SwitchService) ConnectTo(host string, destPort uint32) error {
	// Dial the destination and validate the secret and key
	conn, err := interfaces.SecurityProvider().CanDial(host, destPort)
	if err != nil {
		return err
	}

	config := interfaces.SwitchConfig()
	config.SwitchPort = destPort
	config.Uuid = switchService.switchConfig.Uuid
	config.IsSwitch = true

	err = interfaces.SecurityProvider().ValidateConnection(conn, config)
	if err != nil {
		return err
	}

	config.IsSwitch = false

	edge := edge.NewEdgeImpl(conn, switchService, switchService.registry, switchService.servicePoints, config)

	//Below attributes are only for the port initiating the connection
	/* @TODO implement reconnect between switches
	edge.reconnectInfo = &ReconnectInfo{
		host:         host,
		port:         destPort,
		reconnectMtx: &sync.Mutex{},
	} */

	//We have only one go routing per each because we want to keep the order of incoming and outgoing messages
	edge.Start()
	switchService.notifyNewEdge(edge)
	return nil
}
