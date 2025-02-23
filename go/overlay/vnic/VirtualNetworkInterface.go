package vnic

import (
	"errors"
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/strings"
	"github.com/saichler/shared/go/types"
	"net"
	"sync"
	"time"
)

type VirtualNetworkInterface struct {
	// Resources for this VNic such as registry, security & config
	resources interfaces.IResources
	// The socket connection
	conn net.Conn
	// The socket connection mutex
	connMtx *sync.Mutex
	// is running
	running bool
	// Sub components/go routines
	components *SubComponents
	// The Protocol
	protocol *protocol.Protocol
	// Name for this VNic expressing the connection path in aside -->> zside
	name string
	// Indicates if this vnic in on the switch internal, hence need no keep alive
	IsVNet bool
	// Last reconnect attempt
	last_reconnect_attempt int64
}

func NewVirtualNetworkInterface(resources interfaces.IResources, conn net.Conn) *VirtualNetworkInterface {
	vnic := &VirtualNetworkInterface{}
	vnic.conn = conn
	vnic.resources = resources
	vnic.connMtx = &sync.Mutex{}
	vnic.protocol = protocol.New(resources)
	vnic.components = newSubomponents()
	vnic.components.addComponent(newRX(vnic))
	vnic.components.addComponent(newTX(vnic))
	vnic.resources.Registry().Register(&types.Message{})
	if vnic.resources.Config().LocalUuid == "" {
		vnic.resources.Config().LocalUuid = uuid.New().String()
	}

	if conn == nil {
		// Register the health service
		health.RegisterHealth(vnic.resources, nil)
	}

	return vnic
}

func (vnic *VirtualNetworkInterface) Start() {
	vnic.running = true
	if vnic.conn == nil {
		vnic.resources.Config().ServiceAreas = vnic.resources.ServicePoints().Areas()
		vnic.connectToSwitch()
	} else {
		vnic.receiveConnection()
	}
	vnic.name = vnic.resources.Config().LocalAlias + " -->> " + vnic.resources.Config().RemoteAlias
}

func (vnic *VirtualNetworkInterface) connectToSwitch() {
	vnic.connect()
	vnic.components.start()
}

func (vnic *VirtualNetworkInterface) connect() error {
	// Dial the destination and validate the secret and key
	destination := protocol.MachineIP
	if protocol.UsingContainers {
		// inside a containet the switch ip will be the external subnet + ".1"
		// for example if the address of the container is 172.1.1.112, the switch will be accessible
		// via 172.1.1.1
		subnet := protocol.IpSegment.ExternalSubnet()
		destination = subnet + ".1"
	}
	// Try to dial to the switch
	conn, err := vnic.resources.Security().CanDial(destination, vnic.resources.Config().VnetPort)
	if err != nil {
		return errors.New("Error connecting to the vnet: " + err.Error())
	}
	// Verify that the switch accepts this connection
	err = vnic.resources.Security().ValidateConnection(conn, vnic.resources.Config())
	if err != nil {
		return errors.New("Error validating connection: " + err.Error())
	}
	vnic.conn = conn
	vnic.resources.Config().Address = conn.LocalAddr().String()
	return nil
}

func (vnic *VirtualNetworkInterface) receiveConnection() {
	vnic.IsVNet = true
	vnic.resources.Config().Address = vnic.conn.RemoteAddr().String()
	vnic.components.start()
}

func (vnic *VirtualNetworkInterface) Shutdown() {
	vnic.running = false
	if vnic.conn != nil {
		vnic.conn.Close()
	}
	vnic.components.shutdown()
	if vnic.resources.DataListener() != nil {
		go vnic.resources.DataListener().ShutdownVNic(vnic)
	}
}

func (vnic *VirtualNetworkInterface) Name() string {
	if vnic.name == "" {
		vnic.name = strings.New(vnic.resources.Config().LocalUuid,
			" -->> ",
			vnic.resources.Config().RemoteUuid).String()
	}
	return vnic.name
}

func (vnic *VirtualNetworkInterface) SendMessage(data []byte) error {
	return vnic.components.TX().SendMessage(data)
}

func (vnic *VirtualNetworkInterface) Unicast(action types.Action, destination string, any interface{}) error {
	return vnic.components.TX().Unicast(action, destination, any, 0)
}

func (vnic *VirtualNetworkInterface) Multicast(action types.Action, area int32, topic string, any interface{}) error {
	return vnic.components.TX().Multicast(action, area, topic, any, 0)
}

func (vnic *VirtualNetworkInterface) Resources() interfaces.IResources {
	return vnic.resources
}

func (vnic *VirtualNetworkInterface) reconnect() {
	if !vnic.running {
		return
	}
	vnic.connMtx.Lock()
	defer vnic.connMtx.Unlock()
	if time.Now().Unix()-vnic.last_reconnect_attempt < 5 {
		return
	}
	vnic.last_reconnect_attempt = time.Now().Unix()

	vnic.resources.Logger().Info("***** Trying to reconnect to ", vnic.resources.Config().RemoteAlias, " *****")

	if vnic.conn != nil {
		vnic.conn.Close()
	}

	err := vnic.connect()
	if err != nil {
		vnic.resources.Logger().Error("***** Failed to reconnect to ", vnic.resources.Config().RemoteAlias, " *****")
	} else {
		vnic.resources.Logger().Error("***** Reconnected to ", vnic.resources.Config().RemoteAlias, " *****")
	}
}
