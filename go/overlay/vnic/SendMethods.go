package vnic

import (
	"errors"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"strconv"
)

func (this *VirtualNetworkInterface) Unicast(destination, serviceName string, serviceArea uint16,
	action common.Action, any interface{}) error {
	mobjects := object.New(nil, any)
	return this.components.TX().Unicast(destination, serviceName, serviceArea, action, mobjects, 0,
		false, false, this.protocol.NextMessageNumber(), nil)
}

func (this *VirtualNetworkInterface) Request(destination, serviceName string, serviceArea uint16,
	action common.Action, any interface{}) common.IElements {
	request := this.requests.NewRequest(this.protocol.NextMessageNumber(), this.resources.SysConfig().LocalUuid, 5, this.resources.Logger())

	request.Lock()
	defer request.Unlock()

	var elements common.IElements
	var err error
	query, ok := any.(string)
	if ok {
		elements, err = object.NewQuery(query, this.resources)
		if err != nil {
			return object.NewError(err.Error())
		}
	} else {
		elements = object.New(nil, any)
	}

	e := this.components.TX().Unicast(destination, serviceName, serviceArea, action, elements, 0,
		true, false, request.MsgNum(), nil)
	if e != nil {
		return object.NewError(e.Error())
	}
	request.Wait()
	return request.Response()
}

func (this *VirtualNetworkInterface) Reply(msg common.IMessage, response common.IElements) error {
	reply := msg.(*protocol.Message).ReplyClone(this.resources)
	data, e := this.protocol.CreateMessageForm(reply, response)
	if e != nil {
		this.resources.Logger().Error(e)
		return e
	}
	hc := health.Health(this.resources)
	hp := hc.HealthPoint(msg.Source())
	this.resources.Logger().Debug("Replying to ", msg.Source(), " ", hp.Alias)
	return this.SendMessage(data)
}

func (this *VirtualNetworkInterface) Multicast(serviceName string, serviceArea uint16, action common.Action, any interface{}) error {
	mobjects := object.New(nil, any)
	return this.components.TX().Multicast("", serviceName, serviceArea, action, mobjects, 0,
		false, false, this.protocol.NextMessageNumber(), nil)
}

func (this *VirtualNetworkInterface) Single(serviceName string, serviceArea uint16, action common.Action, any interface{}) (string, error) {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid, false, false)
	if destination == "" {
		return destination, errors.New("Cannot find a destinstion for " + serviceName + " area " +
			strconv.Itoa(int(serviceArea)))
	}

	hp := hc.HealthPoint(destination)
	this.Resources().Logger().Info("Sending Single to ", destination, " alias ", hp.Alias)

	return destination, this.Unicast(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) SingleRequest(serviceName string, serviceArea uint16, action common.Action, any interface{}) common.IElements {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid, false, false)
	if destination == "" {
		return object.NewError("Cannot find a destinstion for " + serviceName + " area " +
			strconv.Itoa(int(serviceArea)))
	}

	hp := hc.HealthPoint(destination)
	this.Resources().Logger().Info("Sending Single Request to ", destination, " alias ", hp.Alias)
	return this.Request(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) Leader(serviceName string, serviceArea uint16, action common.Action, any interface{}) common.IElements {
	hc := health.Health(this.resources)
	destination := hc.DestinationFor(serviceName, serviceArea, this.resources.SysConfig().LocalUuid, false, true)
	return this.Request(destination, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) Forward(msg common.IMessage, destination string) common.IElements {
	pb, err := this.protocol.ElementsOf(msg)
	if err != nil {
		return object.NewError(err.Error())
	}

	request := this.requests.NewRequest(this.protocol.NextMessageNumber(), this.resources.SysConfig().LocalUuid, 5, this.resources.Logger())
	request.Lock()
	defer request.Unlock()

	e := this.components.TX().Unicast(destination, msg.ServiceName(), msg.ServiceArea(), msg.Action(),
		pb, 0, true, false, request.MsgNum(), msg.Tr())
	if e != nil {
		return object.NewError(e.Error())
	}
	request.Wait()
	return request.Response()
}
