package switching

import (
	"github.com/saichler/layer8/go/overlay/edge"
)

func (switchService *SwitchService) ConnectTo(host string, destPort uint32) error {
	sec := switchService.protocol.Providers().Security()
	// Dial the destination and validate the secret and key
	conn, err := sec.CanDial(host, destPort)
	if err != nil {
		return err
	}

	c := switchService.protocol.Providers().Switch()
	config := &c
	config.SwitchPort = destPort
	config.Local_Uuid = switchService.switchConfig.Local_Uuid
	config.IsSwitchSide = true
	config.IsAdjacentASwitch = true

	err = sec.ValidateConnection(conn, config)
	if err != nil {
		return err
	}

	edge := edge.NewEdgeImpl(conn, switchService, nil, config, switchService.protocol.Providers())

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
