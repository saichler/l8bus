package vnic

import (
	"github.com/saichler/types/go/common"
)

type VnicAPI struct {
	serviceName string
	serviceArea int32
	vnic        *VirtualNetworkInterface
	leader      bool
	all         bool
}

func (v VnicAPI) Post(i interface{}) common.IMObjects {

	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Put(i interface{}) common.IMObjects {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Patch(i interface{}) common.IMObjects {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Delete(i interface{}) common.IMObjects {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Get(s string) common.IMObjects {
	//TODO implement me
	panic("implement me")
}

func newAPI(serviceName string, serviceArea int32, leader, all bool) common.API {
	api := &VnicAPI{}
	api.serviceName = serviceName
	api.serviceArea = serviceArea
	api.leader = leader
	api.all = all
	return api
}
