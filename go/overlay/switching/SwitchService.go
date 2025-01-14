package switching

import (
	"errors"
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/state"
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"time"
)

type SwitchService struct {
	switchConfig *types.MessagingConfig
	socket       net.Listener
	active       bool
	ready        bool
	switchTable  *SwitchTable
	protocol     *protocol.Protocol
}

func NewSwitchService(switchConfig *types.MessagingConfig, providers *interfaces.Providers) *SwitchService {
	switchService := &SwitchService{}
	switchService.switchConfig = switchConfig
	switchService.protocol = protocol.New(providers, nil)
	switchService.active = true
	uid, _ := uuid.NewUUID()
	switchService.switchConfig.Local_Uuid = uid.String()
	switchService.switchTable = newSwitchTable(switchService)

	providers.Registry().Register(&types2.States{})
	sp := state.NewStatesServicePoint(providers.Registry(), providers.ServicePoints())
	providers.ServicePoints().RegisterServicePoint(&types2.States{}, sp, providers.Registry())

	return switchService
}

func (switchService *SwitchService) Start() error {
	var err error
	go switchService.start(&err)

	for !switchService.ready && err == nil {
		time.Sleep(time.Millisecond * 50)
	}
	time.Sleep(time.Millisecond * 50)
	return err
}

func (switchService *SwitchService) start(err *error) {
	if switchService.switchConfig.SwitchPort == 0 {
		er := errors.New("Switch Port does not have a port defined")
		err = &er
		return
	}

	er := switchService.bind()
	if er != nil {
		err = &er
		return
	}

	for switchService.active {
		interfaces.Info("Waiting for connections...")
		switchService.ready = true
		conn, e := switchService.socket.Accept()
		if e != nil && switchService.active {
			interfaces.Error("Failed to accept socket connection:", err)
			continue
		}
		if switchService.active {
			interfaces.Info("Accepted socket connection...")
			go switchService.connect(conn)
		}
	}
	interfaces.Warning("Switch Service has ended")
}

func (switchService *SwitchService) bind() error {
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(switchService.switchConfig.SwitchPort)))
	if e != nil {
		return interfaces.Error("Unable to bind to port ", switchService.switchConfig.SwitchPort, e.Error())
	}
	interfaces.Info("Bind Successfully to port ", switchService.switchConfig.SwitchPort)
	switchService.socket = socket
	return nil
}

func (switchService *SwitchService) connect(conn net.Conn) {
	sec := switchService.protocol.Providers().Security()
	err := sec.CanAccept(conn)
	if err != nil {
		interfaces.Error(err)
		return
	}

	edgeC := switchService.protocol.Providers().EdgeSwitch()
	sEdgeConfig := &edgeC
	sEdgeConfig.Local_Uuid = switchService.switchConfig.Local_Uuid
	sEdgeConfig.IsSwitchSide = true

	err = sec.ValidateConnection(conn, sEdgeConfig)
	if err != nil {
		interfaces.Error(err)
		return
	}

	edge := edge.NewEdgeImpl(conn, switchService, nil, sEdgeConfig, switchService.protocol.Providers())
	edge.Start()
	switchService.notifyNewEdge(edge)
}

func (switchService *SwitchService) notifyNewEdge(edge interfaces.IEdge) {
	go switchService.switchTable.addEdge(edge)
}

func (switchService *SwitchService) Shutdown() {
	switchService.active = false
	switchService.socket.Close()
}

func (switchService *SwitchService) HandleData(data []byte, edge interfaces.IEdge) {
	// in case the logger sync is enabled, this will make sure this method logs are grouped together.
	// if the logger sync is not enabled, this will do nothing.
	interfaces.Logger().LoggerLock()
	defer interfaces.Logger().LoggerUnlock()

	interfaces.Trace("********** Swith Service - HandleData **********")
	source, sourceSwitch, destination, _ := protocol.HeaderOf(data)
	interfaces.Trace("** Switch      : ", switchService.switchConfig.Local_Uuid)
	interfaces.Trace("** Source      : ", source)
	interfaces.Trace("** SourceSwitch: ", sourceSwitch)
	interfaces.Trace("** Destination : ", destination)

	//The destination is the switch
	if destination == switchService.switchConfig.Local_Uuid {
		switchService.switchDataReceived(data, edge)
		return
	}

	uuidMap := switchService.switchTable.ServiceUuids(destination, sourceSwitch)
	if uuidMap != nil {
		switchService.sendToPorts(uuidMap, data, sourceSwitch)
		if destination == state.STATE_TOPIC {
			switchService.switchDataReceived(data, edge)
		}
		return
	}

	//The destination is a single port
	_, p := switchService.switchTable.edges.getEdge(destination, switchService.statesServicePoint(), true)
	if p == nil {
		interfaces.Error("Cannot find destination port for ", destination)
		return
	}
	err := p.Send(data)
	if err != nil {
		interfaces.Error("Error sending data:", err)
	}
}

func (switchService *SwitchService) sendToPorts(uuids map[string]bool, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for edgeUuid, _ := range uuids {
		usedUuid, port := switchService.switchTable.edges.getEdge(edgeUuid, switchService.statesServicePoint(),
			switchService.switchConfig.Local_Uuid == sourceSwitch)
		if port != nil {
			// if the port is external, it may already been forward this message
			// so skip it.
			_, ok := alreadySent[usedUuid]
			if !ok {
				alreadySent[usedUuid] = true
				interfaces.Trace("Sending from ", switchService.switchConfig.Local_Uuid, " to ", usedUuid)
				port.Send(data)
			}
		}
	}
}

func (switchService *SwitchService) publish(pb proto.Message) {

}

func (switchService *SwitchService) PortShutdown(edge interfaces.IEdge) {
}

func (switchService *SwitchService) switchDataReceived(data []byte, edge interfaces.IEdge) {
	msg, err := switchService.protocol.MessageOf(data)
	if err != nil {
		interfaces.Error(err)
		return
	}
	pb, err := switchService.protocol.ProtoOf(msg)
	if err != nil {
		interfaces.Error(err)
		return
	}
	// Otherwise call the handler per the action & the type
	interfaces.Info("Switch Service is: ", switchService.switchConfig.Local_Uuid)
	switchService.protocol.Providers().ServicePoints().Handle(pb, msg.Action, edge)
}

func (switchService *SwitchService) Config() types.MessagingConfig {
	return *switchService.switchConfig
}

func (switchService *SwitchService) State() *types2.States {
	return switchService.statesServicePoint().CloneStates()
}

func (switchService *SwitchService) statesServicePoint() *state.StatesServicePoint {
	sp, ok := switchService.protocol.Providers().ServicePoints().ServicePointHandler("States")
	if !ok {
		return nil
	}
	return sp.(*state.StatesServicePoint)
}
