package vnic

import (
	"github.com/saichler/types/go/common"
)

type VnicAPI struct {
	area   int32
	vnic   *VirtualNetworkInterface
	leader bool
	all    bool
}

func (v VnicAPI) Post(i interface{}) (interface{}, error) {

	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Put(i interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Patch(i interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Delete(i interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (v VnicAPI) Get(s string) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func newAPI(area int32, leader, all bool) common.API {
	api := &VnicAPI{}
	api.area = area
	api.leader = leader
	api.all = all
	return api
}
