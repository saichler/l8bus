package vnet

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

type VNet struct {
	resources   interfaces.IResources
	socket      net.Listener
	running     bool
	ready       bool
	switchTable *SwitchTable
	protocol    *protocol.Protocol
}

func NewVNet(resources interfaces.IResources) *VNet {
	net := &VNet{}
	resources.SetDataListener(net)
	net.resources = resources
	net.protocol = protocol.New(resources)
	net.running = true
	net.resources.Config().Local_Uuid = uuid.New().String()
	net.switchTable = newSwitchTable(net)
	health.RegisterHealth(net.resources)
	net.resources.Config().Topics = resources.ServicePoints().Topics()
	return net
}

func (this *VNet) Start() error {
	var err error
	go this.start(&err)

	for !this.ready && err == nil {
		time.Sleep(time.Millisecond * 50)
	}
	time.Sleep(time.Millisecond * 50)
	return err
}

func (this *VNet) start(err *error) {
	if this.resources.Config().SwitchPort == 0 {
		er := errors.New("Switch Port does not have a port defined")
		err = &er
		return
	}

	er := this.bind()
	if er != nil {
		err = &er
		return
	}

	for this.running {
		this.resources.Logger().Info("Waiting for connections...")
		this.ready = true
		conn, e := this.socket.Accept()
		if e != nil && this.running {
			this.resources.Logger().Error("Failed to accept socket connection:", err)
			continue
		}
		if this.running {
			this.resources.Logger().Info("Accepted socket connection...")
			go this.connect(conn)
		}
	}
	this.resources.Logger().Warning("Vnet ", this.resources.Config().LocalAlias, " has ended")
}

func (this *VNet) bind() error {
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(this.resources.Config().SwitchPort)))
	if e != nil {
		return this.resources.Logger().Error("Unable to bind to port ",
			this.resources.Config().SwitchPort, e.Error())
	}
	this.resources.Logger().Info("Bind Successfully to port ",
		this.resources.Config().SwitchPort)
	this.socket = socket
	return nil
}

func (this *VNet) connect(conn net.Conn) {
	sec := this.resources.Security()
	err := sec.CanAccept(conn)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	resources := resources2.NewResources(this.resources.Registry(),
		this.resources.Security(),
		this.resources.ServicePoints(),
		this.resources.Logger(),
		this.resources.Config().LocalAlias)
	resources.SetDataListener(this)

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)
	vnic.Resources().Config().Local_Uuid = this.resources.Config().Local_Uuid

	err = sec.ValidateConnection(conn, vnic.Resources().Config())
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	vnic.Start()
	this.notifyNewVNic(vnic)
}

func (this *VNet) notifyNewVNic(vnic interfaces.IVirtualNetworkInterface) {
	this.switchTable.addVNic(vnic)
}

func (this *VNet) Shutdown() {
	this.running = false
	this.socket.Close()
	this.switchTable.shutdown()
}

func (this *VNet) HandleData(data []byte, edge interfaces.IVirtualNetworkInterface) {
	this.resources.Logger().Trace("********** Swith Service - HandleData **********")
	source, sourceSwitch, destination, _ := protocol.HeaderOf(data)
	this.resources.Logger().Trace("** Switch      : ", this.resources.Config().Local_Uuid)
	this.resources.Logger().Trace("** Source      : ", source)
	this.resources.Logger().Trace("** SourceSwitch: ", sourceSwitch)
	this.resources.Logger().Trace("** Destination : ", destination)

	dSize := len(destination)
	switch dSize {
	case 36:
		//The destination is the switch
		if destination == this.resources.Config().Local_Uuid {
			this.switchDataReceived(data, edge)
			return
		} else {
			//The destination is a single port
			_, p := this.switchTable.conns.getConnection(destination, true, this.resources)
			if p == nil {
				this.resources.Logger().Error("Cannot find destination port for ", destination)
				return
			}
			err := p.Send(data)
			if err != nil {
				this.resources.Logger().Error("Error sending data:", err)
			}
		}
	default:
		uuidMap := this.switchTable.ServiceUuids(destination, sourceSwitch)
		if uuidMap != nil {
			this.sendToPorts(uuidMap, data, sourceSwitch)
			if destination == health.TOPIC {
				this.switchDataReceived(data, edge)
			}
			return
		}
	}
}

func (this *VNet) sendToPorts(uuids map[string]bool, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for edgeUuid, _ := range uuids {
		isHope0 := this.resources.Config().Local_Uuid == sourceSwitch
		usedUuid, port := this.switchTable.conns.getConnection(edgeUuid, isHope0, this.resources)
		if port != nil {
			// if the port is external, it may already been forward this message
			// so skip it.
			_, ok := alreadySent[usedUuid]
			if !ok {
				alreadySent[usedUuid] = true
				this.resources.Logger().Trace("Sending from ", this.resources.Config().Local_Uuid, " to ", usedUuid)
				port.Send(data)
			}
		}
	}
}

func (this *VNet) publish(pb proto.Message) {

}

func (this *VNet) ShutdownVNic(vnic interfaces.IVirtualNetworkInterface) {
}

func (this *VNet) switchDataReceived(data []byte, edge interfaces.IVirtualNetworkInterface) {
	msg, err := this.protocol.MessageOf(data)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}
	pb, err := this.protocol.ProtoOf(msg)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}
	// Otherwise call the handler per the action & the type
	this.resources.Logger().Info("Switch Service is: ", this.resources.Config().Local_Uuid)
	this.resources.ServicePoints().Handle(pb, msg.Action, edge)
}

func (this *VNet) Resources() interfaces.IResources {
	return this.resources
}
