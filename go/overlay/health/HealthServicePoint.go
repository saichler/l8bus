package health

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	types2 "github.com/saichler/shared/go/types"
	"google.golang.org/protobuf/proto"
)

const (
	TOPIC    = "HealthPoint"
	ENDPOINT = "health"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
}

func RegisterHealth(resources interfaces.IResources) {
	health := &HealthServicePoint{}
	health.healthCenter = newHealthCenter(resources)
	ok, err := resources.Registry().Register(&types.HealthPoint{})
	if !ok {
		panic(err)
	}
	err = resources.ServicePoints().RegisterServicePoint(&types.HealthPoint{}, health)
	if err != nil {
		panic(err)
	}
}

func (this *HealthServicePoint) Post(pb proto.Message, vnic interfaces.IVirtualNetworkInterface) (proto.Message, error) {
	hp := pb.(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil, nil
}
func (this *HealthServicePoint) Put(pb proto.Message, vnic interfaces.IVirtualNetworkInterface) (proto.Message, error) {
	hp := pb.(*types.HealthPoint)
	this.healthCenter.Update(hp)
	return nil, nil
}
func (this *HealthServicePoint) Patch(pb proto.Message, vnic interfaces.IVirtualNetworkInterface) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) Delete(pb proto.Message, vnic interfaces.IVirtualNetworkInterface) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) Get(pb proto.Message, vnic interfaces.IVirtualNetworkInterface) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) Failed(pb proto.Message, vnic interfaces.IVirtualNetworkInterface, msg *types2.Message) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) EndPoint() string {
	return ENDPOINT
}
func (this *HealthServicePoint) Topic() string {
	return TOPIC
}
