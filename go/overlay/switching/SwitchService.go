package switching

import (
	"errors"
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	"github.com/saichler/shared/go/share/interfaces"
	resources2 "github.com/saichler/shared/go/share/resources"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"time"
)

type SwitchService struct {
	resources   interfaces.IResources
	socket      net.Listener
	running     bool
	ready       bool
	switchTable *SwitchTable
	protocol    *protocol.Protocol
}

func NewSwitchService(resources interfaces.IResources) *SwitchService {
	switchService := &SwitchService{}
	resources.SetDataListener(switchService)
	switchService.resources = resources
	switchService.protocol = protocol.New(resources)
	switchService.running = true
	switchService.resources.Config().Local_Uuid = uuid.New().String()
	switchService.switchTable = newSwitchTable(switchService)
	health.RegisterHealth(switchService.resources)
	switchService.resources.Config().Topics = resources.ServicePoints().Topics()

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
	if switchService.resources.Config().SwitchPort == 0 {
		er := errors.New("Switch Port does not have a port defined")
		err = &er
		return
	}

	er := switchService.bind()
	if er != nil {
		err = &er
		return
	}

	for switchService.running {
		switchService.resources.Logger().Info("Waiting for connections...")
		switchService.ready = true
		conn, e := switchService.socket.Accept()
		if e != nil && switchService.running {
			switchService.resources.Logger().Error("Failed to accept socket connection:", err)
			continue
		}
		if switchService.running {
			switchService.resources.Logger().Info("Accepted socket connection...")
			go switchService.connect(conn)
		}
	}
	switchService.resources.Logger().Warning("Switch Service has ended")
}

func (switchService *SwitchService) bind() error {
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(switchService.resources.Config().SwitchPort)))
	if e != nil {
		return switchService.resources.Logger().Error("Unable to bind to port ",
			switchService.resources.Config().SwitchPort, e.Error())
	}
	switchService.resources.Logger().Info("Bind Successfully to port ",
		switchService.resources.Config().SwitchPort)
	switchService.socket = socket
	return nil
}

func (switchService *SwitchService) connect(conn net.Conn) {
	sec := switchService.resources.Security()
	err := sec.CanAccept(conn)
	if err != nil {
		switchService.resources.Logger().Error(err)
		return
	}

	resources := resources2.NewResources(switchService.resources.Registry(),
		switchService.resources.Security(),
		switchService.resources.ServicePoints(),
		switchService.resources.Logger(),
		switchService.resources.Config().LocalAlias)
	resources.SetDataListener(switchService)

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)
	vnic.Resources().Config().Local_Uuid = switchService.resources.Config().Local_Uuid

	err = sec.ValidateConnection(conn, vnic.Resources().Config())
	if err != nil {
		switchService.resources.Logger().Error(err)
		return
	}

	vnic.Start()
	switchService.notifyNewVNic(vnic)
}

func (switchService *SwitchService) notifyNewVNic(vnic interfaces.IVirtualNetworkInterface) {
	switchService.switchTable.addVNic(vnic)
}

func (switchService *SwitchService) Shutdown() {
	switchService.running = false
	switchService.socket.Close()
	switchService.switchTable.shutdown()
}

func (switchService *SwitchService) HandleData(data []byte, edge interfaces.IVirtualNetworkInterface) {
	switchService.resources.Logger().Trace("********** Swith Service - HandleData **********")
	source, sourceSwitch, destination, _ := protocol.HeaderOf(data)
	switchService.resources.Logger().Trace("** Switch      : ", switchService.resources.Config().Local_Uuid)
	switchService.resources.Logger().Trace("** Source      : ", source)
	switchService.resources.Logger().Trace("** SourceSwitch: ", sourceSwitch)
	switchService.resources.Logger().Trace("** Destination : ", destination)

	dSize := len(destination)
	switch dSize {
	case 36:
		//The destination is the switch
		if destination == switchService.resources.Config().Local_Uuid {
			switchService.switchDataReceived(data, edge)
			return
		} else {
			//The destination is a single port
			_, p := switchService.switchTable.conns.getConnection(destination, true, switchService.resources)
			if p == nil {
				switchService.resources.Logger().Error("Cannot find destination port for ", destination)
				return
			}
			err := p.Send(data)
			if err != nil {
				switchService.resources.Logger().Error("Error sending data:", err)
			}
		}
	default:
		uuidMap := switchService.switchTable.ServiceUuids(destination, sourceSwitch)
		if uuidMap != nil {
			switchService.sendToPorts(uuidMap, data, sourceSwitch)
			if destination == health.TOPIC {
				switchService.switchDataReceived(data, edge)
			}
			return
		}
	}
}

func (switchService *SwitchService) sendToPorts(uuids map[string]bool, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for edgeUuid, _ := range uuids {
		isHope0 := switchService.resources.Config().Local_Uuid == sourceSwitch
		usedUuid, port := switchService.switchTable.conns.getConnection(edgeUuid, isHope0, switchService.resources)
		if port != nil {
			// if the port is external, it may already been forward this message
			// so skip it.
			_, ok := alreadySent[usedUuid]
			if !ok {
				alreadySent[usedUuid] = true
				switchService.resources.Logger().Trace("Sending from ", switchService.resources.Config().Local_Uuid, " to ", usedUuid)
				port.Send(data)
			}
		}
	}
}

func (switchService *SwitchService) publish(pb proto.Message) {

}

func (switchService *SwitchService) ShutdownVNic(vnic interfaces.IVirtualNetworkInterface) {
}

func (switchService *SwitchService) switchDataReceived(data []byte, edge interfaces.IVirtualNetworkInterface) {
	msg, err := switchService.protocol.MessageOf(data)
	if err != nil {
		switchService.resources.Logger().Error(err)
		return
	}
	pb, err := switchService.protocol.ProtoOf(msg)
	if err != nil {
		switchService.resources.Logger().Error(err)
		return
	}
	// Otherwise call the handler per the action & the type
	switchService.resources.Logger().Info("Switch Service is: ", switchService.resources.Config().Local_Uuid)
	switchService.resources.ServicePoints().Handle(pb, msg.Action, edge)
}

func (switchService *SwitchService) Resources() interfaces.IResources {
	return switchService.resources
}
