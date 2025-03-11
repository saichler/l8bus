package vnic

import (
	"errors"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/reflect/go/reflect/cloning"
	"github.com/saichler/shared/go/share/strings"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"net"
	"sync"
	"time"
)

type VirtualNetworkInterface struct {
	// Resources for this VNic such as registry, security & config
	resources common.IResources
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

	requests *Requests

	stats *types.HealthPointStats
}

func NewVirtualNetworkInterface(resources common.IResources, conn net.Conn) *VirtualNetworkInterface {
	vnic := &VirtualNetworkInterface{}
	vnic.conn = conn
	vnic.resources = resources
	vnic.connMtx = &sync.Mutex{}
	vnic.protocol = protocol.New(resources)
	vnic.components = newSubomponents()
	vnic.components.addComponent(newRX(vnic))
	vnic.components.addComponent(newTX(vnic))
	vnic.components.addComponent(newKeepAlive(vnic))
	vnic.requests = newRequests()
	vnic.resources.Registry().Register(&types.Message{})
	vnic.resources.Registry().Register(&types.Transaction{})
	vnic.stats = &types.HealthPointStats{}
	if vnic.resources.Config().LocalUuid == "" {
		vnic.resources.Config().LocalUuid = common.NewUuid()
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
	err = vnic.resources.Security().ValidateConnection(conn)
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
	return vnic.components.TX().Unicast(action, destination, any, 0, false, false, vnic.protocol.NextMessageNumber(), nil)
}

func (vnic *VirtualNetworkInterface) Multicast(cast types.CastMode, action types.Action, vlan int32, topic string, any interface{}) error {
	return vnic.components.TX().Multicast(action, vlan, topic, any, 0, false, false, vnic.protocol.NextMessageNumber(), nil)
}

func (vnic *VirtualNetworkInterface) Request(cast types.CastMode, action types.Action, vlan int32, topic string, any interface{}) (interface{}, error) {
	hc := health.Health(vnic.resources)
	destination := hc.UuidsForRequest(cast, vlan, topic, vnic.resources.Config().LocalUuid)

	request := vnic.requests.newRequest(vnic.protocol.NextMessageNumber(), vnic.resources.Config().LocalUuid)
	request.cond.L.Lock()
	defer request.cond.L.Unlock()

	e := vnic.components.TX().Unicast(action, destination, any, 0, true, false, request.msgNum, nil)
	if e != nil {
		return nil, e
	}
	request.cond.Wait()
	return request.response, nil
}

func (vnic *VirtualNetworkInterface) Transaction(action types.Action, vlan int32, topic string, any interface{}) (interface{}, error) {
	return vnic.Request(types.CastMode_Single, action, vlan, topic, any)
}

func (vnic *VirtualNetworkInterface) Reply(msg *types.Message, resp interface{}) error {
	reply := cloning.NewCloner().Clone(msg).(*types.Message)
	reply.Action = types.Action_Reply
	reply.Topic = msg.SourceUuid
	reply.SourceUuid = vnic.resources.Config().LocalUuid
	reply.SourceVnetUuid = vnic.resources.Config().RemoteUuid
	reply.IsRequest = false
	reply.IsReply = true

	data, e := vnic.protocol.CreateMessageForm(reply, resp)
	if e != nil {
		vnic.resources.Logger().Error(e)
		return e
	}
	return vnic.SendMessage(data)
}

func (vnic *VirtualNetworkInterface) Forward(msg *types.Message, destination string) (interface{}, error) {
	pb, err := vnic.protocol.ProtoOf(msg)
	if err != nil {
		return nil, err
	}

	request := vnic.requests.newRequest(vnic.protocol.NextMessageNumber(), vnic.resources.Config().LocalUuid)
	request.cond.L.Lock()
	defer request.cond.L.Unlock()

	e := vnic.components.TX().Unicast(msg.Action, destination, pb, 0, true, false, request.msgNum, msg.Tr)
	if e != nil {
		return nil, e
	}
	request.cond.Wait()
	return request.response, nil
}

func (vnic *VirtualNetworkInterface) API(vlan int32) common.API {
	return newAPI(vlan, types.CastMode_Leader)
}

func (vnic *VirtualNetworkInterface) Resources() common.IResources {
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
