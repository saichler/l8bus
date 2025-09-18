package vnic

import (
	"reflect"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
	"google.golang.org/protobuf/proto"
)

func (this *VirtualNetworkInterface) Forward(msg *ifs.Message, destination string) ifs.IElements {
	pb, err := this.protocol.ElementsOf(msg)
	if err != nil {
		return object.NewError(err.Error())
	}

	timeout := 15
	if msg.Tr_Timeout() > 0 {
		timeout = int(msg.Tr_Timeout())
	}

	request := this.requests.NewRequest(this.protocol.NextMessageNumber(), this.resources.SysConfig().LocalUuid, timeout, this.resources.Logger())
	defer this.requests.DelRequest(request.MsgNum(), request.MsgSource())

	request.Lock()
	defer request.Unlock()

	e := this.components.TX().Unicast(destination, msg.ServiceName(), msg.ServiceArea(), msg.Action(),
		pb, ifs.P8, ifs.M_All, true, false, request.MsgNum(),
		msg.Tr_State(), msg.Tr_Id(), msg.Tr_ErrMsg(), msg.Tr_StartTime(), msg.Tr_Timeout(), msg.AAAId())
	if e != nil {
		return object.NewError(e.Error())
	}
	request.Wait()
	return request.Response()
}

func createElements(any interface{}, resources ifs.IResources) (ifs.IElements, error) {
	if any == nil {
		return object.New(nil, nil), nil
	}
	pq, ok := any.(*types.Query)
	if ok {
		return object.NewQuery(pq.Text, resources)
	}

	gsql, ok := any.(string)
	if ok {
		return object.NewQuery(gsql, resources)
	}

	elems, ok := any.(ifs.IElements)
	if ok {
		return elems, nil
	}

	pb, ok := any.(proto.Message)
	if ok {
		return object.New(nil, pb), nil
	}

	v := reflect.ValueOf(any)

	if v.Kind() == reflect.Slice {
		pbs := make([]proto.Message, v.Len())
		for i := 0; i < v.Len(); i++ {
			elm := v.Index(i)
			pb, ok = elm.Interface().(proto.Message)
			if ok {
				pbs[i] = pb
			} else {
				panic(strings.New("Uknown input type ", reflect.ValueOf(pb).String()).String())
			}
		}
		return object.New(nil, pbs), nil
	}
	panic(strings.New("Uknown input type ", reflect.ValueOf(any).String()).String())
}
