package vnet

import (
	"time"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8types/go/types/l8system"
)

func (this *VNet) PropertyChangeNotification(set *l8notify.L8NotificationSet) {
	//only health service will call this callback so check if the notification is from a local source
	//if it is from local source, then just notify local vnics
	protocol.AddPropertyChangeCalled(set, this.resources.SysConfig().LocalAlias)
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()
	syncData, _ := this.protocol.CreateMessageFor("", set.ServiceName, byte(set.ServiceArea), ifs.P1, ifs.M_All,
		ifs.Notify, vnetUuid, vnetUuid, object.New(nil, set), false, false,
		nextId, ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")

	go this.HandleData(syncData, nil)
}

func (this *VNet) publisLocalHealth() {
	time.Sleep(time.Second)
	vnetUuid := this.resources.SysConfig().LocalUuid
	vnetName := this.resources.SysConfig().LocalAlias

	local := this.switchTable.conns.allInternals()
	ext := this.switchTable.conns.allExternals()

	this.resources.Logger().Debug("Vnet ", vnetName, " publish health ", len(local), " ext ", len(ext))

	if len(local) > 0 {
		hps := make([]*l8health.L8Health, 0)
		for uuid, _ := range local {
			hp := health.HealthOf(uuid, this.resources)
			hps = append(hps, hp)
		}
		localHealth := health.HealthOf(vnetUuid, this.resources)
		hps = append(hps, localHealth)
		for extUuid, _ := range ext {
			extHealth := health.HealthOf(extUuid, this.resources)
			hps = append(hps, extHealth)
		}

		nextId := this.protocol.NextMessageNumber()
		sync, _ := this.protocol.CreateMessageFor("", health.ServiceName, health.ServiceArea, ifs.P1, ifs.M_All,
			ifs.PATCH, vnetUuid, vnetUuid, object.New(nil, hps), false, false,
			nextId, ifs.NotATransaction, "", "",
			-1, -1, -1, -1, -1, 0, false, "")
		for _, conn := range ext {
			conn.SendMessage(sync)
		}
	}
}

func (this *VNet) publishRoutes() {
	vnetUuid := this.resources.SysConfig().LocalUuid
	vnetName := this.resources.SysConfig().LocalAlias

	nextId := this.protocol.NextMessageNumber()

	routeTable := &l8system.L8RouteTable{Rows: this.switchTable.conns.Routes()}
	this.resources.Logger().Debug("Vnet ", vnetName, " publish routes ", len(routeTable.Rows))

	data := &l8system.L8SystemMessage_RouteTable{RouteTable: routeTable}
	routes := &l8system.L8SystemMessage{Action: l8system.L8SystemAction_Routes_Add, Data: data}

	routesData, _ := this.protocol.CreateMessageFor("", ifs.SysMsg, ifs.SysArea, ifs.P1, ifs.M_All,
		ifs.POST, vnetUuid, vnetUuid, object.New(nil, routes), false, false,
		nextId, ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")

	allExternal := this.switchTable.conns.allExternals()
	for _, external := range allExternal {
		external.SendMessage(routesData)
	}
	go this.publisLocalHealth()
}

func (this *VNet) publishRemovedRoutes(removed map[string]string) {
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()

	routeTable := &l8system.L8RouteTable{Rows: removed}
	data := &l8system.L8SystemMessage_RouteTable{RouteTable: routeTable}
	routes := &l8system.L8SystemMessage{Action: l8system.L8SystemAction_Routes_Remove, Data: data}

	routesData, _ := this.protocol.CreateMessageFor("", ifs.SysMsg, ifs.SysArea, ifs.P1, ifs.M_All,
		ifs.POST, vnetUuid, vnetUuid, object.New(nil, routes), false, false,
		nextId, ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")

	allExternal := this.switchTable.conns.allExternals()
	for _, external := range allExternal {
		external.SendMessage(routesData)
	}
}

func (this *VNet) publishSystemMessage(sysmsg *l8system.L8SystemMessage) {
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()

	sysmsg.Publish = false

	sysmsgData, _ := this.protocol.CreateMessageFor("", ifs.SysMsg, ifs.SysArea, ifs.P1, ifs.M_All,
		ifs.POST, vnetUuid, vnetUuid, object.New(nil, sysmsg), false, false,
		nextId, ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")

	allExternal := this.switchTable.conns.allExternals()
	for _, external := range allExternal {
		external.SendMessage(sysmsgData)
	}
}
