package vnet

import (
	"errors"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	resources2 "github.com/saichler/l8utils/go/utils/resources"
	"github.com/saichler/l8utils/go/utils/strings"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"time"
)

type VNet struct {
	resources   ifs.IResources
	socket      net.Listener
	running     bool
	ready       bool
	switchTable *SwitchTable
	protocol    *protocol.Protocol
	discovery   *Discovery
	ns          *NotificationSender
}

func NewVNet(resources ifs.IResources) *VNet {
	net := &VNet{}
	net.resources = resources
	net.resources.Set(net)

	net.protocol = protocol.New(net.resources)
	net.running = true
	net.resources.SysConfig().LocalUuid = ifs.NewUuid()
	net.switchTable = newSwitchTable(net)

	net.resources.Services().RegisterServiceHandlerType(&health.HealthService{})
	net.resources.Services().Activate(health.ServiceTypeName, health.ServiceName, 0, net.resources, net)
	net.discovery = NewDiscovery(net)
	net.ns = newNotificationSender(net)

	hc := health.Health(net.resources)
	hp := &types.Health{}
	hp.Alias = net.resources.SysConfig().LocalAlias
	hp.AUuid = net.resources.SysConfig().LocalUuid
	hp.IsVnet = true
	hp.Services = net.resources.SysConfig().Services
	hc.Add(hp, true)
	return net
}

func (this *VNet) Start() error {
	var err error
	go this.start(&err)
	for !this.ready && err == nil {
		time.Sleep(time.Millisecond * 50)
	}
	time.Sleep(time.Millisecond * 50)
	this.discovery.Discover()
	return err
}

func (this *VNet) start(err *error) {
	this.resources.Logger().Info("*** VNET Start Ver 1.0")
	if this.resources.SysConfig().VnetPort == 0 {
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
	this.resources.Logger().Warning("Vnet ", this.resources.SysConfig().LocalAlias, " has ended")
}

func (this *VNet) bind() error {
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(this.resources.SysConfig().VnetPort)))
	if e != nil {
		return this.resources.Logger().Error("Unable to bind to port ",
			this.resources.SysConfig().VnetPort, e.Error())
	}
	this.resources.Logger().Info("Bind Successfully to port ",
		this.resources.SysConfig().VnetPort)
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

	config := &types.SysConfig{MaxDataSize: resources2.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize: resources2.DEFAULT_QUEUE_SIZE,
		TxQueueSize: resources2.DEFAULT_QUEUE_SIZE,
		LocalAlias:  this.resources.SysConfig().LocalAlias,
		LocalUuid:   this.resources.SysConfig().LocalUuid,
		Services: &types.Services{ServiceToAreas: map[string]*types.ServiceAreas{
			health.ServiceName: &types.ServiceAreas{
				Areas: map[int32]bool{0: true},
			}}},
	}

	resources := resources2.NewResources(this.resources.Logger())
	resources.Copy(this.resources)
	resources.Set(this)
	resources.Set(config)

	vnic := vnic2.NewVirtualNetworkInterface(resources, conn)
	vnic.Resources().SysConfig().LocalUuid = this.resources.SysConfig().LocalUuid

	err = sec.ValidateConnection(conn, config)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	vnic.Start()
	this.notifyNewVNic(vnic)
}

func (this *VNet) notifyNewVNic(vnic ifs.IVNic) {
	this.switchTable.addVNic(vnic)
}

func (this *VNet) Shutdown() {
	this.resources.Logger().Info("Shutdown called!")
	this.running = false
	this.socket.Close()
	this.switchTable.shutdown()
}

func (this *VNet) Failed(data []byte, vnic ifs.IVNic, failMsg string) {
	msg, err := this.protocol.MessageOf(data, this.resources)
	this.resources.Logger().Error("Failed Message ", msg.Action, ":", failMsg)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	failMessage := msg.CloneFail(failMsg, this.resources.SysConfig().RemoteUuid)
	data, _ = failMessage.Marshal(nil, this.resources)

	err = vnic.SendMessage(data)
	if err != nil {
		this.resources.Logger().Error(err)
	}
}

func (this *VNet) HandleData(data []byte, vnic ifs.IVNic) {
	this.resources.Logger().Trace("********** Swith Service - HandleData **********")
	source, sourceVnet, destination, serviceName, serviceArea, _ := ifs.HeaderOf(data)
	this.resources.Logger().Trace("** Switch       : ", this.resources.SysConfig().LocalUuid)
	this.resources.Logger().Trace("** Source       : ", source)
	this.resources.Logger().Trace("** SourceVnet   : ", sourceVnet)
	this.resources.Logger().Trace("** Destination  : ", destination)
	this.resources.Logger().Trace("** Service Name : ", serviceName)
	this.resources.Logger().Trace("** Service Area : ", serviceArea)

	if destination != "" {
		//The destination is the vnet
		if destination == this.resources.SysConfig().LocalUuid {
			this.switchDataReceived(data, vnic)
			return
		}
		if destination == ifs.DESTINATION_Single {
			uuidMap := this.switchTable.ServiceUuids(serviceName, serviceArea, sourceVnet)
			if uuidMap != nil {
				for uuid, _ := range uuidMap {
					if uuid != source {
						destination = uuid
						break
					}
				}
			}
		}
		//The destination is a single port
		_, p := this.switchTable.conns.getConnection(destination, true, this.resources)
		if p == nil {
			this.resources.Logger().Error("Unknown destination ", destination)
			this.Failed(data, vnic, strings.New("Cannot find destination port for ", destination).String())
			return
		}
		err := p.SendMessage(data)
		if err != nil {
			this.resources.Logger().Error("Unable to send to destination ", destination)
			this.Failed(data, vnic, strings.New("Error sending data:", err.Error()).String())
			return
		}

	} else {
		uuidMap := this.switchTable.ServiceUuids(serviceName, serviceArea, sourceVnet)
		if uuidMap != nil {
			this.uniCastToPorts(uuidMap, data, sourceVnet)
			if serviceName == health.ServiceName && source != this.resources.SysConfig().LocalUuid {
				this.switchDataReceived(data, vnic)
			}
			return
		}
	}
}

func (this *VNet) uniCastToPorts(uuids map[string]bool, data []byte, sourceSwitch string) {
	alreadySent := make(map[string]bool)
	for vnicUuid, _ := range uuids {
		isHope0 := this.resources.SysConfig().LocalUuid == sourceSwitch
		usedUuid, port := this.switchTable.conns.getConnection(vnicUuid, isHope0, this.resources)
		if port != nil {
			// if the port is external, it may already been forward this message
			// so skip it.
			_, ok := alreadySent[usedUuid]
			if !ok {
				alreadySent[usedUuid] = true
				this.resources.Logger().Trace("Sending from ", this.resources.SysConfig().LocalUuid, " to ", usedUuid)
				port.SendMessage(data)
			}
		}
	}
}

func (this *VNet) publish(pb proto.Message) {

}

func (this *VNet) ShutdownVNic(vnic ifs.IVNic) {
	h := health.Health(this.resources)
	uuid := vnic.Resources().SysConfig().RemoteUuid
	hp := h.Health(uuid)
	if hp.Status != types.HealthState_Down {
		hp.Status = types.HealthState_Down
		h.Update(hp, false)
	}
	this.resources.Logger().Info("Shutdown complete ", vnic.Resources().SysConfig().LocalAlias)
}

func (this *VNet) switchDataReceived(data []byte, vnic ifs.IVNic) {
	msg, err := this.protocol.MessageOf(data, this.resources)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	pb, err := this.protocol.ElementsOf(msg)
	if err != nil {
		if msg.Tr_State() != ifs.Empty {
			//This message should not be processed and we should just
			//reply with nil to unblock the transaction
			vnic.Reply(msg, nil)
			return
		}
		this.resources.Logger().Error(err)
		return
	}
	// Otherwise call the handler per the action & the type
	this.resources.Logger().Trace("Switch Service is: ", this.resources.SysConfig().LocalUuid)
	if msg.Action() == ifs.Notify {
		resp := this.resources.Services().Notify(pb, vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			panic(pb)
			this.resources.Logger().Error(resp.Error())
		}
	} else {
		resp := this.resources.Services().Handle(pb, msg.Action(), vnic, msg)
		if resp != nil && resp.Error() != nil {
			this.resources.Logger().Error(resp.Error())
		}
	}
}

func (this *VNet) Resources() ifs.IResources {
	return this.resources
}

func (this *VNet) LocalCount() int {
	return len(this.switchTable.conns.internal)
}

func (this *VNet) ExternalCount() int {
	return len(this.switchTable.conns.external)
}
