package switching

import (
	"errors"
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/state"
	types2 "github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	logs "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"time"
)

type SwitchService struct {
	switchConfig  *types.MessagingConfig
	socket        net.Listener
	active        bool
	ready         bool
	switchTable   *SwitchTable
	registry      interfaces.IStructRegistry
	servicePoints interfaces.IServicePoints
}

func NewSwitchService(switchConfig *types.MessagingConfig, registry interfaces.IStructRegistry, servicePoints interfaces.IServicePoints) *SwitchService {
	switchService := &SwitchService{}
	switchService.switchConfig = switchConfig
	switchService.servicePoints = servicePoints
	switchService.active = true
	switchService.registry = registry
	uid, _ := uuid.NewUUID()
	switchService.switchConfig.Local_Uuid = uid.String()
	switchService.switchTable = newSwitchTable(switchService)

	registry.RegisterStruct(&types2.States{})
	sp := state.NewStatesServicePoint(registry, servicePoints)
	servicePoints.RegisterServicePoint(&types2.States{}, sp, registry)

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
		logs.Info("Waiting for connections...")
		switchService.ready = true
		conn, e := switchService.socket.Accept()
		if e != nil && switchService.active {
			logs.Error("Failed to accept socket connection:", err)
			continue
		}
		if switchService.active {
			logs.Info("Accepted socket connection...")
			go switchService.connect(conn)
		}
	}
	logs.Warning("Switch Service has ended")
}

func (switchService *SwitchService) bind() error {
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(switchService.switchConfig.SwitchPort)))
	if e != nil {
		return logs.Error("Unable to bind to port ", switchService.switchConfig.SwitchPort, e.Error())
	}
	logs.Info("Bind Successfully to port ", switchService.switchConfig.SwitchPort)
	switchService.socket = socket
	return nil
}

func (switchService *SwitchService) connect(conn net.Conn) {
	err := interfaces.SecurityProvider().CanAccept(conn)
	if err != nil {
		logs.Error(err)
		return
	}

	sEdgeConfig := interfaces.EdgeSwitchConfig()
	sEdgeConfig.Local_Uuid = switchService.switchConfig.Local_Uuid
	sEdgeConfig.IsSwitchSide = true

	err = interfaces.SecurityProvider().ValidateConnection(conn, sEdgeConfig)
	if err != nil {
		logs.Error(err)
		return
	}

	edge := edge.NewEdgeImpl(conn, switchService, switchService.registry, nil, sEdgeConfig)
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
	logs.Trace("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
	source, sourceSwitch, destination, _ := protocol.HeaderOf(data)
	logs.Trace("Switch      : ", switchService.switchConfig.Local_Uuid)
	logs.Trace("Source      : ", source)
	logs.Trace("SourceSwitch: ", sourceSwitch)
	logs.Trace("Destination : ", destination)

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
	_, p := switchService.switchTable.edges.getEdge(destination, "", false)
	if p == nil {
		logs.Error("Cannot find destination port for ", destination)
		return
	}
	p.Send(data)
}

func (switchService *SwitchService) sendToPorts(uuids map[string]string, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for edgeUuid, remoteUuid := range uuids {
		usedUuid, port := switchService.switchTable.edges.getEdge(edgeUuid, remoteUuid,
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
	msg, err := protocol.MessageOf(data)
	if err != nil {
		logs.Error(err)
		return
	}
	pb, err := protocol.ProtoOf(msg, switchService.registry)
	if err != nil {
		logs.Error(err)
		return
	}
	// Otherwise call the handler per the action & the type
	logs.Info("Switch Service is: ", switchService.switchConfig.Local_Uuid)
	switchService.servicePoints.Handle(pb, msg.Action, edge)
}

func (switchService *SwitchService) Config() types.MessagingConfig {
	return *switchService.switchConfig
}

func (switchService *SwitchService) State() {
	switchService.statesServicePoint().CloneStates()
}

func (switchService *SwitchService) statesServicePoint() *state.StatesServicePoint {
	sp, ok := switchService.servicePoints.ServicePointHandler("States")
	if !ok {
		return nil
	}
	return sp.(*state.StatesServicePoint)
}
