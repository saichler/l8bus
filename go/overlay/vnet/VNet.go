package vnet

import (
	"errors"
	"net"
	"time"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	resources2 "github.com/saichler/l8utils/go/utils/resources"
	"github.com/saichler/l8utils/go/utils/strings"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
)

type VNet struct {
	resources   ifs.IResources
	socket      net.Listener
	running     bool
	ready       bool
	switchTable *SwitchTable
	protocol    *protocol.Protocol
	discovery   *Discovery
}

func NewVNet(resources ifs.IResources) *VNet {
	resources.Registry().Register(&types.SystemMessage{})
	resources.Registry().Register(&types.Empty{})
	resources.Registry().Register(&types.Top{})
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

	hc := health.Health(net.resources)
	hp := &types.Health{}
	hp.Alias = net.resources.SysConfig().LocalAlias
	hp.AUuid = net.resources.SysConfig().LocalUuid
	hp.IsVnet = true
	hp.Services = net.resources.SysConfig().Services
	hc.Put(hp, true)
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
	this.resources.Logger().Info("VNet.start: Starting VNet ")
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
	socket, e := net.Listen("tcp", strings.New(":", int(this.resources.SysConfig().VnetPort)).String())
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
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	failMessage := msg.CloneFail(failMsg, this.resources.SysConfig().RemoteUuid)
	data, _ = failMessage.Marshal(nil, this.resources)

	err = vnic.SendMessage(data)
}

func (this *VNet) HandleData(data []byte, vnic ifs.IVNic) {
	protocol.AddHandleData()
	source, sourceVnet, destination, serviceName, serviceArea, _, multicastMode := ifs.HeaderOf(data)

	if serviceName == ifs.SysMsg && serviceArea == ifs.SysArea {
		go this.systemMessageReceived(data, vnic)
		return
	}

	if destination != "" {
		//The destination is the vnet
		if destination == this.resources.SysConfig().LocalUuid {
			go this.vnetServiceRequest(data, vnic)
			return
		}
		if destination == ifs.DESTINATION_Single {
			destination = this.switchTable.services.serviceFor(serviceName, serviceArea, source, multicastMode)
		}
		//The destination is a single port
		_, p := this.switchTable.conns.getConnection(destination, true)
		if p == nil {
			this.Failed(data, vnic, strings.New("Cannot find destination port for ", destination).String())
			return
		}

		err := p.SendMessage(data)
		if err != nil {
			if !p.Running() {
				uuid := p.Resources().SysConfig().RemoteUuid
				h := health.Health(this.resources)
				hp := h.Health(uuid)
				this.sendHealth(hp)
			}
			this.Failed(data, vnic, strings.New("Error sending data:", err.Error()).String())
			return
		}
	} else {
		connections := this.switchTable.connectionsForService(serviceName, serviceArea, sourceVnet, multicastMode)
		this.uniCastToPorts(connections, data)
		if serviceName == health.ServiceName && source != this.resources.SysConfig().LocalUuid {
			go this.vnetServiceRequest(data, vnic)
		}
		return
	}
}

func (this *VNet) uniCastToPorts(connections map[string]ifs.IVNic, data []byte) {
	for _, port := range connections {
		port.SendMessage(data)
	}
}

func (this *VNet) ShutdownVNic(vnic ifs.IVNic) {
	uuid := vnic.Resources().SysConfig().RemoteUuid
	removed := map[string]string{uuid: ""}
	this.switchTable.routeTable.removeRoutes(removed)
	this.switchTable.services.removeService(removed)
	this.removeHealth(removed)
	this.publishRemovedRoutes(removed)
}

func (this *VNet) Resources() ifs.IResources {
	return this.resources
}

func (this *VNet) requestHealthSync() {
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()
	sync, _ := this.protocol.CreateMessageFor("", health.ServiceName, 0, ifs.P1, ifs.M_All,
		ifs.Sync, vnetUuid, vnetUuid, object.New(nil, nil), false, false,
		nextId, ifs.Empty, "", "", -1, -1, "")
	go this.HandleData(sync, nil)
}

func (this *VNet) sendHealth(hp *types.Health) {
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()
	h, _ := this.protocol.CreateMessageFor("", health.ServiceName, 0, ifs.P1, ifs.M_All,
		ifs.POST, vnetUuid, vnetUuid, object.New(nil, hp), false, false,
		nextId, ifs.Empty, "", "", -1, -1, "")
	go this.HandleData(h, nil)
}
