package switching

import (
	"errors"
	"fmt"
	"github.com/saichler/overlayK8s/go/edge"
	"github.com/saichler/overlayK8s/go/protocol"
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
	switchService.switchTable = newSwitchTable()
	switchService.active = true
	switchService.registry = registry
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
	sEdgeConfig.Uuid = switchService.switchConfig.Uuid
	sEdgeConfig.ZUuid, err = interfaces.SecurityProvider().ValidateConnection(conn, sEdgeConfig.Uuid)
	if err != nil {
		logs.Error(err)
		return
	}
	sEdgeConfig.IsSwitch = true

	edge := edge.NewEdgeImpl(conn, switchService, switchService.registry, switchService.servicePoints, sEdgeConfig)
	edge.Start()
	switchService.notifyNewEdge(edge)
}

func (switchService *SwitchService) notifyNewEdge(edge interfaces.IEdge) {
	go switchService.switchTable.addEdge(edge, switchService.switchConfig.Uuid)
}

func (switchService *SwitchService) Shutdown() {
	switchService.active = false
	switchService.socket.Close()
}

func (switchService *SwitchService) HandleData(data []byte, edge interfaces.IEdge) {
	source, destination, pri := protocol.HeaderOf(data)
	fmt.Println(source, destination, pri.String())
	//The destination is the switch
	if destination == switchService.switchConfig.Uuid {
		switchService.switchDataReceived(data, edge)
		return
	}

	uuidList := switchService.switchTable.stateCenter.ServiceUuids(destination)
	if uuidList != nil {
		switchService.sendToPorts(uuidList, data)
		return
	}

	//The destination is a single port
	p := switchService.switchTable.fetchEdgeByUuid(destination)
	if p == nil {
		logs.Error("Cannot find destination port for ", destination)
		return
	}
	p.Send(data)
}

func (switchService *SwitchService) sendToPorts(uuids []string, data []byte) {
	for _, uuid := range uuids {
		port := switchService.switchTable.fetchEdgeByUuid(uuid)
		if port != nil {
			port.Send(data)
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
	switchService.servicePoints.Handle(pb, msg.Request.Type, edge)
}
