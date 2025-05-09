package vnic

import (
	"errors"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/l8utils/go/utils/strings"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/plugins"
	"github.com/saichler/layer8/go/overlay/protocol"
	requests2 "github.com/saichler/layer8/go/overlay/vnic/requests"
	"net"
	"os"
	"sync"
	"time"
)

type VirtualNetworkInterface struct {
	// Resources for this VNic such as registry, security & config
	resources ifs.IResources
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

	requests *requests2.Requests

	stats     *types.HealthStats
	connected bool
}

func NewVirtualNetworkInterface(resources ifs.IResources, conn net.Conn) *VirtualNetworkInterface {
	vnic := &VirtualNetworkInterface{}
	vnic.conn = conn
	vnic.resources = resources
	vnic.connMtx = &sync.Mutex{}
	vnic.protocol = protocol.New(resources)
	vnic.components = newSubomponents()
	vnic.components.addComponent(newRX(vnic))
	vnic.components.addComponent(newTX(vnic))
	vnic.components.addComponent(newKeepAlive(vnic))
	vnic.requests = requests2.NewRequests()
	vnic.stats = &types.HealthStats{}
	if vnic.resources.SysConfig().LocalUuid == "" {
		vnic.resources.SysConfig().LocalUuid = ifs.NewUuid()
	}

	if conn == nil {
		// Register the health service
		vnic.resources.Services().RegisterServiceHandlerType(&health.HealthService{})
		vnic.resources.Services().Activate(health.ServiceTypeName, health.ServiceName, 0, vnic.resources, nil)
		vnic.resources.Services().RegisterServiceHandlerType(&plugins.PluginService{})
		h, _ := vnic.resources.Services().Activate(plugins.ServiceTypeName, plugins.ServiceName, 0, vnic.resources, nil)
		h.(*plugins.PluginService).Vnic = vnic
	}

	return vnic
}

func (this *VirtualNetworkInterface) Start() {
	this.running = true
	if this.conn == nil {
		this.connectToSwitch()
	} else {
		this.receiveConnection()
	}
	this.name = this.resources.SysConfig().LocalAlias + " -->> " + this.resources.SysConfig().RemoteAlias
}

func (this *VirtualNetworkInterface) connectToSwitch() {
	err := this.connect()
	if err != nil {
		panic(err)
	}
	this.components.start()
	this.connected = true
}

func (this *VirtualNetworkInterface) connect() error {
	// Dial the destination and validate the secret and key
	destination := protocol.MachineIP
	if ifs.NetworkMode_K8s() {
		destination = os.Getenv("NODE_IP")
	} else if ifs.NetworkMode_DOCKER() {
		// inside a containet the switch ip will be the external subnet + ".1"
		// for example if the address of the container is 172.1.1.112, the switch will be accessible
		// via 172.1.1.1
		subnet := protocol.IpSegment.ExternalSubnet()
		destination = subnet + ".1"

	}
	this.resources.Logger().Info("Trying to connect to vnet at IP - ", destination)
	// Try to dial to the switch
	conn, err := this.resources.Security().CanDial(destination, this.resources.SysConfig().VnetPort)
	if err != nil {
		return errors.New("Error connecting to the vnet: " + err.Error())
	}
	// Verify that the switch accepts this connection
	if this.resources.SysConfig().LocalUuid == "" {
		panic("Couldn't connect")
	}
	err = this.resources.Security().ValidateConnection(conn, this.resources.SysConfig())
	if err != nil {
		return errors.New("Error validating connection: " + err.Error())
	}
	this.conn = conn
	this.resources.SysConfig().Address = conn.LocalAddr().String()
	this.resources.Logger().Info("Connected!")
	return nil
}

func (this *VirtualNetworkInterface) receiveConnection() {
	this.IsVNet = true
	this.resources.SysConfig().Address = this.conn.RemoteAddr().String()
	this.components.start()
}

func (this *VirtualNetworkInterface) Shutdown() {
	this.resources.Logger().Info("Shutdown was called")
	this.running = false
	if this.conn != nil {
		this.conn.Close()
	}
	this.components.shutdown()
	if this.resources.DataListener() != nil {
		go this.resources.DataListener().ShutdownVNic(this)
	}
}

func (this *VirtualNetworkInterface) Name() string {
	if this.name == "" {
		this.name = strings.New(this.resources.SysConfig().LocalUuid,
			" -->> ",
			this.resources.SysConfig().RemoteUuid).String()
	}
	return this.name
}

func (this *VirtualNetworkInterface) SendMessage(data []byte) error {
	return this.components.TX().SendMessage(data)
}

func (this *VirtualNetworkInterface) ServiceAPI(serviceName string, serviceArea uint16) ifs.ServiceAPI {
	return newAPI(serviceName, serviceArea, false, false)
}

func (this *VirtualNetworkInterface) Resources() ifs.IResources {
	return this.resources
}

func (this *VirtualNetworkInterface) reconnect() {
	if !this.running {
		return
	}
	this.connMtx.Lock()
	defer this.connMtx.Unlock()
	if time.Now().Unix()-this.last_reconnect_attempt < 5 {
		return
	}
	this.last_reconnect_attempt = time.Now().Unix()

	this.resources.Logger().Info("***** Trying to reconnect to ", this.resources.SysConfig().RemoteAlias, " *****")

	if this.conn != nil {
		this.conn.Close()
	}

	err := this.connect()
	if err != nil {
		this.resources.Logger().Error("***** Failed to reconnect to ", this.resources.SysConfig().RemoteAlias, " *****")
	} else {
		this.resources.Logger().Info("***** Reconnected to ", this.resources.SysConfig().RemoteAlias, " *****")
	}
}

func (this *VirtualNetworkInterface) Stats() *types.HealthStats {
	return this.stats
}

func (this *VirtualNetworkInterface) WaitForConnection() {
	for !this.connected {
		time.Sleep(time.Millisecond * 100)
	}
	hc := health.Health(this.resources)
	hp := hc.Health(this.resources.SysConfig().LocalUuid)
	for hp == nil {
		time.Sleep(time.Millisecond * 100)
		hp = hc.Health(this.resources.SysConfig().LocalUuid)
	}
}
