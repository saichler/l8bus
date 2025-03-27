package vnic

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/reflect/go/reflect/cloning"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
)

func (this *VirtualNetworkInterface) Unicast(destination, serviceName string, serviceArea int32,
	action types.Action, any interface{}) error {
	mobjects := object.New("", any)
	return this.components.TX().Unicast(destination, serviceName, serviceArea, action, mobjects, 0,
		false, false, this.protocol.NextMessageNumber(), nil)
}

func (this *VirtualNetworkInterface) Request(destination, serviceName string, serviceArea int32,
	action types.Action, any interface{}) common.IMObjects {
	request := this.requests.newRequest(this.protocol.NextMessageNumber(), this.resources.Config().LocalUuid)
	request.cond.L.Lock()
	defer request.cond.L.Unlock()
	mobjects := object.New("", any)
	e := this.components.TX().Unicast(destination, serviceName, serviceArea, action, mobjects, 0,
		true, false, request.msgNum, nil)
	if e != nil {
		return object.NewError(e.Error())
	}
	request.cond.Wait()
	return request.response
}

func (this *VirtualNetworkInterface) Reply(msg *types.Message, response common.IMObjects) error {
	reply := cloning.NewCloner().Clone(msg).(*types.Message)
	reply.Action = types.Action_Reply
	reply.Destination = msg.Source
	reply.Source = this.resources.Config().LocalUuid
	reply.SourceVnet = this.resources.Config().RemoteUuid
	reply.IsRequest = false
	reply.IsReply = true

	data, e := this.protocol.CreateMessageForm(reply, response)
	if e != nil {
		this.resources.Logger().Error(e)
		return e
	}
	return this.SendMessage(data)
}

func (this *VirtualNetworkInterface) Multicast(serviceName string, serviceArea int32, action types.Action, any interface{}) error {
	mobjects := object.New("", any)
	return this.components.TX().Multicast("", serviceName, serviceArea, action, mobjects, 0,
		false, false, this.protocol.NextMessageNumber(), nil)
}

func (this *VirtualNetworkInterface) Single(serviceName string, serviceArea int32, action types.Action, any interface{}) error {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.Config().LocalUuid, false, false)
	return this.Unicast(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) SingleRequest(serviceName string, serviceArea int32, action types.Action, any interface{}) common.IMObjects {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.Config().LocalUuid, false, false)
	return this.Request(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) Leader(serviceName string, serviceArea int32, action types.Action, any interface{}) common.IMObjects {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.Config().LocalUuid, false, true)
	return this.Request(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) Forward(msg *types.Message, destination string) common.IMObjects {
	pb, err := this.protocol.MObjectsOf(msg)
	if err != nil {
		return object.NewError(err.Error())
	}

	request := this.requests.newRequest(this.protocol.NextMessageNumber(), this.resources.Config().LocalUuid)
	request.cond.L.Lock()
	defer request.cond.L.Unlock()

	e := this.components.TX().Unicast(destination, msg.ServiceName, msg.ServiceArea, msg.Action,
		pb, 0, true, false, request.msgNum, msg.Tr)
	if e != nil {
		return object.NewError(e.Error())
	}
	request.cond.Wait()
	return request.response
}
