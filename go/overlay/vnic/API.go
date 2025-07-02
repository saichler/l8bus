package vnic

import (
	"github.com/saichler/l8types/go/ifs"
)

type VnicAPI struct {
	serviceName string
	serviceArea byte
	vnic        *VirtualNetworkInterface
	leader      bool
	all         bool
}

func (v VnicAPI) Post(i interface{}) ifs.IElements {

	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Put(i interface{}) ifs.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Patch(i interface{}) ifs.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Delete(i interface{}) ifs.IElements {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Get(s string) ifs.IElements {
	//TODO implement me
	panic("implement me")
}

func newAPI(serviceName string, serviceArea byte, leader, all bool) ifs.ServiceAPI {
	api := &VnicAPI{}
	api.serviceName = serviceName
	api.serviceArea = serviceArea
	api.leader = leader
	api.all = all
	return api
}
