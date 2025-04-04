package vnic

import (
	"github.com/saichler/types/go/common"
)

type VnicAPI struct {
	serviceName string
	serviceArea uint16
	vnic        *VirtualNetworkInterface
	leader      bool
	all         bool
}

func (v VnicAPI) Post(i interface{}) common.IElements {

	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Put(i interface{}) common.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Patch(i interface{}) common.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Delete(i interface{}) common.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Get(s string) common.IElements {
	//TODO implement me
	panic("implement me")
}

func newAPI(serviceName string, serviceArea uint16, leader, all bool) common.API {
	api := &VnicAPI{}
	api.serviceName = serviceName
	api.serviceArea = serviceArea
	api.leader = leader
	api.all = all
	return api
}
