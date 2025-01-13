package edge

import (
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/protocol"
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
	serializer interfaces.Serializer,
	config *types.MessagingConfig,
	providers *interfaces.Providers) *EdgeImpl {

	edge := &EdgeImpl{}
	edge.config = config
	edge.protocol = protocol.New(providers, serializer)
	edge.createdAtTime = time.Now().Unix()
	edge.conn = con
	edge.connMtx = &sync.Mutex{}
	edge.active = true
	edge.dataListener = dataListener

	if edge.config.IsSwitchSide {
		edge.config.Address = con.RemoteAddr().String()
	} else {
		edge.config.Address = con.LocalAddr().String()
	}

	edge.rx = queues.NewByteSliceQueue("RX", int(config.RxQueueSize))
	edge.tx = queues.NewByteSliceQueue("TX", int(config.TxQueueSize))

	registry := providers.Registry()
	registry.Register(&types.Message{})

	_, err := registry.TypeInfo("States")
	// If there is an error, this servicepoints already registered so do nothing
	if err != nil {
		registry.Register(&types2.States{})
		sp := state.NewStatesServicePoint(registry, providers.ServicePoints())
		providers.ServicePoints().RegisterServicePoint(&types2.States{}, sp, registry)
		edge.localState = state.CreateStatesFromConfig(edge.config, true)
	}
	return edge
}

// This is the method that the service port is using to connect to the switch for the VM/machine
func ConnectTo(host string,
	destPort uint32,
	datalistener interfaces.IDatatListener,
	serializer interfaces.Serializer,
	config *types.MessagingConfig,
	providers *interfaces.Providers) (interfaces.IEdge, error) {

	// Dial the destination and validate the secret and key
	conn, err := providers.Security().CanDial(host, destPort)
	if err != nil {
		return nil, err
	}

	config.Local_Uuid = uuid.New().String()
	config.IsSwitchSide = false

	err = providers.Security().ValidateConnection(conn, config)
	if err != nil {
		return nil, err
	}

	edge := newEdgeImpl(conn, datalistener, serializer, config, providers)

	//Below attributes are only for the port initiating the connection
	edge.reconnectInfo = &ReconnectInfo{
		host:         host,
		port:         destPort,
		reconnectMtx: &sync.Mutex{},
	}

	//Update the switch uuid in the list of services, if this is an edge connection
	edge.updateRemoteUuid()

	//We have only one go routing per each because we want to keep the order of incoming and outgoing messages
	edge.Start()

	return edge, nil
}

func NewEdgeImpl(
	con net.Conn,
	dataListener interfaces.IDatatListener,
	serializer interfaces.Serializer,
	config *types.MessagingConfig,
	providers *interfaces.Providers) *EdgeImpl {
	return newEdgeImpl(con, dataListener, serializer, config, providers)
}

func (edge *EdgeImpl) SetName(name string) {
	edge.name = name
}
