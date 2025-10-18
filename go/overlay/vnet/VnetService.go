package vnet

import "github.com/saichler/l8types/go/ifs"

func (this *VNet) vnetServiceRequest(data []byte, vnic ifs.IVNic) {
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

	// Otherwise call the handler per the action & the type
	if msg.Action() == ifs.Notify {
		resp := this.resources.Services().Notify(pb, vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			panic(pb)
			this.resources.Logger().Error(resp.Error())
		}
		return
	}

	if msg.Reply() {
		this.vnic.SetResponse(msg, pb)
		return
	}

	resp := this.resources.Services().Handle(pb, msg.Action(), vnic, msg)
	if resp != nil && resp.Error() != nil {
		this.resources.Logger().Error(resp.Error(), " : ", msg.Action())
	}
	if msg.Request() {
		err = vnic.Reply(msg, resp)
		if err != nil {
			this.resources.Logger().Error(err.Error())
		}
	}
}

func (this *VNet) ExternalCount() int32 {
	return this.switchTable.conns.sizeExternal.Load()
}

func (this *VNet) LocalCount() int32 {
	return this.switchTable.conns.sizeInternal.Load()
}
