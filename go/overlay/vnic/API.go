package vnic

import (
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
)

type VnicAPI struct {
	area int32
	cast types.CastMode
	vnic *VirtualNetworkInterface
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

func newAPI(area int32, cast types.CastMode) interfaces.API {
	api := &VnicAPI{}
	api.area = area
	api.cast = cast
	return api
}
