package edge

import (
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/state"
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/queues"
	"github.com/saichler/shared/go/types"
	"net"
	"sync"
	"time"
)

// NewEdgeImpl Instantiate a new port with a connection
func newEdgeImpl(
	con net.Conn,
	dataListener interfaces.IDatatListener,
	registry interfaces.IStructRegistry,
	servicePoints interfaces.IServicePoints,
	config *types.MessagingConfig) *EdgeImpl {

	edge := &EdgeImpl{}
	edge.config = config
	edge.registry = registry
	edge.servicePoints = servicePoints
	edge.createdAtTime = time.Now().Unix()
	edge.conn = con
	edge.active = true
	edge.dataListener = dataListener

	if edge.config.IsSwitchSide {
		edge.config.Address = con.RemoteAddr().String()
	} else {
		edge.config.Address = con.LocalAddr().String()
	}

	edge.rx = queues.NewByteSliceQueue("RX", int(config.RxQueueSize))
	edge.tx = queues.NewByteSliceQueue("TX", int(config.TxQueueSize))

	edge.registry.RegisterStruct(&types2.States{})
	edge.stateServicePoint = state.NewStatesServicePoint(edge.registry, edge.servicePoints)
	edge.servicePoints.RegisterServicePoint(&types2.States{}, edge.stateServicePoint, edge.registry)

	return edge
}

// This is the method that the service port is using to connect to the switch for the VM/machine
func ConnectTo(host string,
	destPort uint32,
	datalistener interfaces.IDatatListener,
	registry interfaces.IStructRegistry,
	servicePoints interfaces.IServicePoints,
	config *types.MessagingConfig) (interfaces.IEdge, error) {

	// Dial the destination and validate the secret and key
	conn, err := interfaces.SecurityProvider().CanDial(host, destPort)
	if err != nil {
		return nil, err
	}

	config.Local_Uuid = uuid.New().String()
	config.IsSwitchSide = false

	err = interfaces.SecurityProvider().ValidateConnection(conn, config)
	if err != nil {
		return nil, err
	}

	edge := newEdgeImpl(conn, datalistener, registry, servicePoints, config)

	//Below attributes are only for the port initiating the connection
	edge.reconnectInfo = &ReconnectInfo{
		host:         host,
		port:         destPort,
		reconnectMtx: &sync.Mutex{},
	}

	//We have only one go routing per each because we want to keep the order of incoming and outgoing messages
	edge.Start()

	return edge, nil
}

func NewEdgeImpl(
	con net.Conn,
	dataListener interfaces.IDatatListener,
	registry interfaces.IStructRegistry,
	servicePoints interfaces.IServicePoints,
	config *types.MessagingConfig) *EdgeImpl {
	return newEdgeImpl(con, dataListener, registry, servicePoints, config)
}
