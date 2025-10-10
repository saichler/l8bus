package vnic

import (
	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
)

func (this *VirtualNetworkInterface) Unicast(destination, serviceName string, serviceArea byte,
	action ifs.Action, any interface{}) error {
	return this.unicast(destination, serviceName, serviceArea, action, any, ifs.P8, ifs.M_All)
}

func (this *VirtualNetworkInterface) unicast(destination, serviceName string, serviceArea byte,
	action ifs.Action, any interface{}, priority ifs.Priority, multicastMode ifs.MulticastMode) error {

	if destination == "" {
		destination = ifs.DESTINATION_Single
	}

	elems, err := createElements(any, this.resources)
	if err != nil {
		return err
	}
	return this.components.TX().Unicast(destination, serviceName, serviceArea, action, elems, priority, multicastMode,
		false, false, this.protocol.NextMessageNumber(), ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, "")
}

func (this *VirtualNetworkInterface) Request(destination, serviceName string, serviceArea byte,
	action ifs.Action, any interface{}, timeoutSeconds int, tokens ...string) ifs.IElements {
	return this.request(destination, serviceName, serviceArea, action, any, ifs.P8, ifs.M_All, timeoutSeconds, tokens...)
}

func (this *VirtualNetworkInterface) request(destination, serviceName string, serviceArea byte,
	action ifs.Action, any interface{}, priority ifs.Priority, multicastMode ifs.MulticastMode, timeoutInSeconds int, tokens ...string) ifs.IElements {

	if destination == "" {
		destination = ifs.DESTINATION_Single
	}

	request := this.requests.NewRequest(this.protocol.NextMessageNumber(), this.resources.SysConfig().LocalUuid, timeoutInSeconds, this.resources.Logger())
	defer this.requests.DelRequest(request.MsgNum(), request.MsgSource())

	request.Lock()
	defer request.Unlock()

	elements, err := createElements(any, this.resources)
	if err != nil {
		return object.NewError(err.Error())
	}
	token := ""
	if tokens != nil && len(tokens) > 0 {
		token = tokens[0]
	}
	e := this.components.TX().Unicast(destination, serviceName, serviceArea, action, elements, priority, multicastMode,
		true, false, request.MsgNum(), ifs.NotATransaction, "", "",
		-1, -1, -1, -1, int64(timeoutInSeconds), 0, token)
	if e != nil {
		return object.NewError(e.Error())
	}
	request.Wait()
	return request.Response()
}

func (this *VirtualNetworkInterface) Reply(msg *ifs.Message, response ifs.IElements) error {
	reply := msg.CloneReply(this.resources.SysConfig().LocalUuid, this.resources.SysConfig().RemoteUuid)
	data, e := this.protocol.CreateMessageForm(reply, response)
	if e != nil {
		this.resources.Logger().Error(e)
		return e
	}
	hc := health.Health(this.resources)
	hp := hc.Health(msg.Source())
	alias := " No Alias Yet"
	if hp != nil {
		alias = hp.Alias
	}
	this.resources.Logger().Debug("Replying to ", msg.Source(), " ", alias)
	return this.SendMessage(data)
}
