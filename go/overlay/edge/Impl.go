package edge

import (
	"github.com/saichler/overlayK8s/go/overlay/state"
	"github.com/saichler/shared/go/share/interfaces"
	//This is just to not put interfaces.Debug for example
	log "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/queues"
	strutil "github.com/saichler/shared/go/share/string_utils"
	"github.com/saichler/shared/go/types"
	"net"
	"sync"
	"time"
)

type EdgeImpl struct {
	// The config for this edge, including uuids.
	config *types.MessagingConfig
	// The incoming data queue
	rx *queues.ByteSliceQueue
	// The outgoing data queue
	tx *queues.ByteSliceQueue
	// The connection
	conn net.Conn
	// is the port active
	active bool
	// The incoming data listener
	dataListener interfaces.IDatatListener
	// The used registry
	registry interfaces.IStructRegistry
	// Service Points
	servicePoints interfaces.IServicePoints
	//edge reconnect info, only valid if the port is the initiating side
	reconnectInfo *ReconnectInfo
	// created at
	createdAtTime int64
	// Last Message Sent
	lastMessageSentTime int64
	// Pre-prepared status request to clone from
	status *types.Status
}

type ReconnectInfo struct {
	//The host
	host string
	//The port
	port uint32
	// Mutex as multiple go routines might call reconnect
	reconnectMtx *sync.Mutex
	// Indicates if the port was already reconnected
	alreadyReconnected bool
}

func (edge *EdgeImpl) Start() {
	// Start loop reading from the socket
	go edge.readFromSocket()
	// Start loop reading from the TX queue and writing to the socket
	go edge.writeToSocket()
	// Start loop notifying the raw data listener on new incoming data
	go edge.notifyRawDataListener()

	go edge.reportStatus()
	log.Info(edge.Name(), "Started!")
}

func (edge *EdgeImpl) Config() types.MessagingConfig {
	return *edge.config
}

func (edge *EdgeImpl) Shutdown() {
	log.Info(edge.Name(), "Shutdown called...")
	edge.active = false
	if edge.conn != nil {
		edge.conn.Close()
	}
	edge.rx.Shutdown()
	edge.tx.Shutdown()

	if edge.dataListener != nil {
		edge.dataListener.PortShutdown(edge)
	}
}

func (edge *EdgeImpl) attemptToReconnect() {
	// Should not be a valid scenario, however bugs do happen
	if edge.reconnectInfo == nil {
		return
	}

	edge.reconnectInfo.reconnectMtx.Lock()
	defer edge.reconnectInfo.reconnectMtx.Unlock()
	if edge.reconnectInfo.alreadyReconnected {
		return
	}
	for {
		time.Sleep(time.Second * 5)
		log.Warning("Connection issues, trying to reconnect to switch")

		err := edge.reconnect()
		if err == nil {
			edge.reconnectInfo.alreadyReconnected = true
			go func() {
				time.Sleep(time.Second)
				edge.reconnectInfo.alreadyReconnected = false
			}()
			break
		}

	}
	log.Info("Reconnected!")
}

func (edge *EdgeImpl) reconnect() error {
	// Dial the destination and validate the secret and key
	conn, err := interfaces.SecurityProvider().CanDial(edge.reconnectInfo.host, edge.reconnectInfo.port)
	if err != nil {
		return err
	}
	err = interfaces.SecurityProvider().ValidateConnection(conn, edge.config)
	if err != nil {
		return err
	}

	edge.conn = conn

	return nil
}

func (edge *EdgeImpl) Name() string {
	name := strutil.New("")
	if edge.config.IsSwitch {
		name.Add("Switch Port ")
	} else {
		name.Add("Node Port ")
	}
	name.Add(edge.config.Uuid)
	name.Add("[")
	name.Add(edge.config.Addr)
	name.Add("]")
	return name.String()
}

func (edge *EdgeImpl) CreatedAt() int64 {
	return edge.createdAtTime
}

func (edge *EdgeImpl) reportStatus() {
	for edge.active {
		time.Sleep(time.Second * 5)
		if time.Now().Unix() > edge.lastMessageSentTime+5 {
			edge.lastMessageSentTime = time.Now().Unix()
			request := &types.Request{Type: types.Action_POST, Status: edge.status}
			edge.Do(request, state.STATE_TOPIC, nil)
		}
	}
}
