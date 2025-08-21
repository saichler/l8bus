package vnet

import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/layer8/go/overlay/protocol"
)

func (this *VNet) PropertyChangeNotification(set *types.NotificationSet) {
	//only health service will call this callback so check if the notification is from a local source
	//if it is from local source, then just notify local vnics
	protocol.AddPropertyChangeCalled(set, this.resources.SysConfig().LocalAlias)
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()
	syncData, _ := this.protocol.CreateMessageFor("", set.ServiceName, byte(set.ServiceArea), ifs.P1, ifs.M_All,
		ifs.Notify, vnetUuid, vnetUuid, object.New(nil, set), false, false,
		nextId, ifs.Empty, "", "", -1, "")

	go this.HandleData(syncData, nil)
}

func (this *VNet) publishRoutes() {
	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()

	routeTable := &types.RouteTable{Rows: this.switchTable.conns.Routes()}
	data := &types.SystemMessage_RouteTable{RouteTable: routeTable}
	routes := &types.SystemMessage{Action: types.SystemAction_Routes_Add, Data: data}

	routesData, _ := this.protocol.CreateMessageFor("", ifs.SysMsg, ifs.SysArea, ifs.P1, ifs.M_All,
		ifs.POST, vnetUuid, vnetUuid, object.New(nil, routes), false, false,
		nextId, ifs.Empty, "", "", -1, "")

	allExternal := this.switchTable.conns.allExternals()
	for _, external := range allExternal {
		external.SendMessage(routesData)
	}
}
