package vnet

import (
	"errors"
	"github.com/google/uuid"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	"github.com/saichler/shared/go/share/interfaces"
	resources2 "github.com/saichler/shared/go/share/resources"
	"github.com/saichler/shared/go/share/strings"
	"github.com/saichler/shared/go/types"
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
	net.resources = resources2.NewResources(resources.Registry(),
		resources.Security(),
		resources.ServicePoints(),
		resources.Logger(),
		net,
		resources.Serializer(interfaces.BINARY), resources.Config(),
		resources.Introspector())
	net.protocol = protocol.New(net.resources)
	net.running = true
	net.resources.Config().LocalUuid = uuid.New().String()
	net.switchTable = newSwitchTable(net)
	net.resources.Config().ServiceAreas = net.resources.ServicePoints().Areas()
	health.RegisterHealth(net.resources, net)
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
	if this.resources.Config().VnetPort == 0 {
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
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(this.resources.Config().VnetPort)))
	if e != nil {
		return this.resources.Logger().Error("Unable to bind to port ",
			this.resources.Config().VnetPort, e.Error())
	}
	this.resources.Logger().Info("Bind Successfully to port ",
		this.resources.Config().VnetPort)
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

	config := &types.VNicConfig{MaxDataSize: resources2.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize: resources2.DEFAULT_QUEUE_SIZE,
		TxQueueSize: resources2.DEFAULT_QUEUE_SIZE,
		LocalAlias:  this.resources.Config().LocalAlias,
		LocalUuid:   this.resources.Config().LocalUuid}

	resources := resources2.NewResources(this.resources.Registry(),
		this.resources.Security(),
		this.resources.ServicePoints(),
		this.resources.Logger(),
		this,
		this.resources.Serializer(interfaces.BINARY),
		config,
		this.resources.Introspector())

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)
	vnic.Resources().Config().LocalUuid = this.resources.Config().LocalUuid

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

func (this *VNet) Failed(data []byte, vnic interfaces.IVirtualNetworkInterface, failMsg string) {
	msg, err := this.protocol.MessageOf(data)
	this.resources.Logger().Error("Failed Message ", msg.Action)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}
	msg.FailMsg = failMsg
	src := msg.SourceUuid
	msg.SourceUuid = msg.Topic
	msg.Topic = src
	data, err = this.protocol.DataFromMessage(msg)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}
	err = vnic.SendMessage(data)
	if err != nil {
		this.resources.Logger().Error(err)
	}
}

func (this *VNet) HandleData(data []byte, vnic interfaces.IVirtualNetworkInterface) {
	this.resources.Logger().Trace("********** Swith Service - HandleData **********")
	source, sourceVnet, destination, area, _ := protocol.HeaderOf(data)
	this.resources.Logger().Trace("** Switch      : ", this.resources.Config().LocalUuid)
	this.resources.Logger().Trace("** Source      : ", source)
	this.resources.Logger().Trace("** SourceVnet: ", sourceVnet)
	this.resources.Logger().Trace("** Area      : ", area)
	this.resources.Logger().Trace("** Destination : ", destination)

	dSize := len(destination)
	switch dSize {
	case protocol.UNICAST_ADDRESS_SIZE:
		//The destination is the vnet
		if destination == this.resources.Config().LocalUuid {
			this.switchDataReceived(data, vnic)
			return
		} else {
			//The destination is a single port
			_, p := this.switchTable.conns.getConnection(destination, true, this.resources)
			if p == nil {
				this.Failed(data, vnic, strings.New("Cannot find destination port for ", destination).String())
				return
			}
			err := p.SendMessage(data)
			if err != nil {
				this.Failed(data, vnic, strings.New("Error sending data:", err.Error()).String())
				return
			}
		}
	default:
		uuidMap := this.switchTable.ServiceUuids(area, destination, sourceVnet)
		if uuidMap != nil {
			this.uniCastToPorts(uuidMap, data, sourceVnet)
			if destination == health.TOPIC {
				this.switchDataReceived(data, vnic)
			}
			return
		}
	}
}

func (this *VNet) uniCastToPorts(uuids map[string]bool, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for vnicUuid, _ := range uuids {
		isHope0 := this.resources.Config().LocalUuid == sourceSwitch
		usedUuid, port := this.switchTable.conns.getConnection(vnicUuid, isHope0, this.resources)
		if port != nil {
			// if the port is external, it may already been forward this message
			// so skip it.
			_, ok := alreadySent[usedUuid]
			if !ok {
				alreadySent[usedUuid] = true
				this.resources.Logger().Trace("Sending from ", this.resources.Config().LocalUuid, " to ", usedUuid)
				port.SendMessage(data)
			}
		}
	}
}

func (this *VNet) publish(pb proto.Message) {

}

func (this *VNet) ShutdownVNic(vnic interfaces.IVirtualNetworkInterface) {
	h := health.Health(this.resources)
	uuid := vnic.Resources().Config().RemoteUuid
	hp := h.GetHealthPoint(uuid)
	if hp.Status != types.HealthState_Down {
		hp.Status = types.HealthState_Down
		h.Update(hp)
		//this.resources.Logger().Trace(this.resources.Config().LocalAlias, " Updated health state: ", hp.Alias, " to ", hp.Status)
		//this.switchTable.sendToAll(health.TOPIC, types.Action_PUT, hp)
	}
	this.resources.Logger().Info("Shutdown complete ", this.resources.Config().LocalAlias)
}

func (this *VNet) switchDataReceived(data []byte, vnic interfaces.IVirtualNetworkInterface) {
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
	this.resources.Logger().Info("Switch Service is: ", this.resources.Config().LocalUuid)
	this.resources.ServicePoints().Handle(pb, msg.Action, vnic, msg)
}

func (this *VNet) Resources() interfaces.IResources {
	return this.resources
}

func (this *VNet) PropertyChangeNotification(set *types.NotificationSet) {
	this.switchTable.uniCastToAll(0, set.TypeName, types.Action_Notify, set)
}
