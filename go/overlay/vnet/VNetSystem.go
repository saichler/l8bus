package vnet

import (
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

func (this *VNet) systemMessageReceived(data []byte, vnic ifs.IVNic) {
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

	systemMessage := pb.Element().(*types.SystemMessage)

	if systemMessage.Action == types.SystemAction_Routes_Add {
		added := this.switchTable.conns.addRoutes(systemMessage.GetRouteTable().Rows)
		this.requestTop(added)
		return
	}
}

func (this *VNet) requestTop(added map[string]string) {
	if len(added) > 0 {
		this.resources.Logger().Info("Route Table ", this.resources.SysConfig().VnetPort, " Size is:", this.switchTable.conns.RouteTableSize(), " added ", len(added))
		this.publishRoutes()
		this.requestHealthSync()
	}
}
