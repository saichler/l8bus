package edge

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/state"
	types2 "github.com/saichler/layer8/go/types"
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
	// Conn mtx
	connMtx *sync.Mutex
	// is the port active
	active bool
	// The incoming data listener
	dataListener interfaces.IDatatListener
	// The used registry
	registry interfaces.ITypeRegistry
	// Service Points
	servicePoints interfaces.IServicePoints
	//edge reconnect info, only valid if the port is the initiating side
	reconnectInfo *ReconnectInfo
	// created at
	createdAtTime int64
	// Last Message Sent
	lastMessageSentTime int64

	name       string
	localState *types2.States
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

func (edge *EdgeImpl) servicePoint() *state.StatesServicePoint {
	if edge.servicePoints == nil {
		return nil
	}
	sp, ok := edge.servicePoints.ServicePointHandler("States")
	if !ok {
		return nil
	}
	return sp.(*state.StatesServicePoint)
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
	if edge.name != "" {
		return edge.name
	}
	name := strutil.New("")
	if edge.config.IsSwitchSide {
		name.Add("Switch Port ")
	} else {
		name.Add("Node Port ")
	}
	name.Add(edge.config.Local_Uuid)
	name.Add("->")
	name.Add(edge.config.RemoteUuid)
	name.Add("[")
	name.Add(edge.config.Address)
	name.Add("]")
	return name.String()
}

func (edge *EdgeImpl) CreatedAt() int64 {
	return edge.createdAtTime
}

func (edge *EdgeImpl) reportStatus() {
	if !edge.config.SendStateInfo {
		return
	}
	for edge.active {
		time.Sleep(time.Second * time.Duration(edge.config.SendStateIntervalSeconds))
		if time.Now().Unix() > edge.lastMessageSentTime+edge.config.SendStateIntervalSeconds {
			edge.lastMessageSentTime = time.Now().Unix()
			if edge.localState == nil {
				break
			}
			edge.PublishState()
		}
	}
}

func (edge *EdgeImpl) State() *types2.States {
	return edge.servicePoint().CloneStates()
}

func (edge *EdgeImpl) PublishState() {
	data, err := protocol.CreateMessageFor(types.Priority_P0, types.Action_POST,
		edge.config.Local_Uuid, edge.config.RemoteUuid,
		edge.servicePoint().Topic(), edge.localState, edge.registry)
	if err != nil {
		log.Error("Failed to create state message: ", err)
		return
	}
	err = edge.Send(data)
	if err != nil {
		log.Error("Failed to send state: ", err)
		return
	}
}
