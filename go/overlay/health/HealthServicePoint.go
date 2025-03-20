package health

import (
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
	typ          *reflect.Type
}

func RegisterHealth(resources common.IResources, listener cache.ICacheListener) {
	health := &HealthServicePoint{}
	health.healthCenter = newHealthCenter(resources, listener)
	err := resources.ServicePoints().RegisterServicePoint(Multicast, 0, health)
	if err != nil {
		panic(err)
	}
}

func (this *HealthServicePoint) Post(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	hp := pb.(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil, nil
}
func (this *HealthServicePoint) Put(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	hp := pb.(*types.HealthPoint)
	this.healthCenter.Update(hp)
	return nil, nil
}
func (this *HealthServicePoint) Patch(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	hp := pb.(*types.HealthPoint)
	this.healthCenter.Update(hp)
	return nil, nil
}
func (this *HealthServicePoint) Delete(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) GetCopy(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) Get(pb proto.Message, resourcs common.IResources) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) Failed(pb proto.Message, resourcs common.IResources, msg *types.Message) (proto.Message, error) {
	return nil, nil
}
func (this *HealthServicePoint) EndPoint() string {
	return Endpoint
}
func (this *HealthServicePoint) Multicast() string {
	return Multicast
}
func (this *HealthServicePoint) Transactional() bool {
	return false
}

func (this *HealthServicePoint) ReplicationCount() int {
	return 0
}

func (this *HealthServicePoint) ReplicationScore() int {
	return 0
}

func (this *HealthServicePoint) SupportedProto() proto.Message {
	return &types.HealthPoint{}
}
