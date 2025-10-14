package vnet

import (
	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8system"
)

func (this *VNet) systemMessageReceived(data []byte, vnic ifs.IVNic) {
	msg, err := this.protocol.MessageOf(data, this.resources)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	pb, err := this.protocol.ElementsOf(msg)
	if err != nil {
		if msg.Tr_State() != ifs.NotATransaction {
			//This message should not be processed and we should just
			//reply with nil to unblock the transaction
			vnic.Reply(msg, nil)
			return
		}
		this.resources.Logger().Error(err)
		return
	}

	systemMessage := pb.Element().(*l8system.L8SystemMessage)

	switch systemMessage.Action {
	case l8system.L8SystemAction_Routes_Add:
		added := this.switchTable.routeTable.addRoutes(systemMessage.GetRouteTable().Rows)
		this.routesAdded(added)
		return
	case l8system.L8SystemAction_Routes_Remove:
		removed := this.switchTable.routeTable.removeRoutes(systemMessage.GetRouteTable().Rows)
		this.routesRemoved(removed)
		return
	case l8system.L8SystemAction_Service_Add:
		serviceData := systemMessage.GetServiceData()
		this.switchTable.services.addService(serviceData)
		if systemMessage.Publish {
			this.publishSystemMessage(systemMessage)
			//go health.AddServiceToHealth(msg.Source(), serviceData.ServiceName, serviceData.ServiceArea, this.resources)
		}
		return
	default:
		panic("unknown system action")
	}
}

func (this *VNet) routesAdded(added map[string]string) {
	if len(added) > 0 {
		this.publishRoutes()
	}
}

func (this *VNet) routesRemoved(removed map[string]string) {
	if len(removed) > 0 {
		this.switchTable.services.removeService(removed)
		this.publishRemovedRoutes(removed)
		this.removeHealth(removed)
	}
}

func (this *VNet) removeHealth(removed map[string]string) {
	hs, _ := health.HealthService(this.resources)
	for uuid, _ := range removed {
		hp := health.HealthOf(uuid, this.resources)
		if hp != nil {
			hs.Delete(object.New(nil, hp), this.vnic)
		}
	}
}
