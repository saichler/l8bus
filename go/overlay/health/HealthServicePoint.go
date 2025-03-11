package health

import (
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"google.golang.org/protobuf/proto"
)

const (
	TOPIC    = "HealthPoint"
	ENDPOINT = "health"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
}

func RegisterHealth(resources common.IResources, listener cache.ICacheListener) {
	health := &HealthServicePoint{}
	health.healthCenter = newHealthCenter(resources, listener)
	err := resources.ServicePoints().RegisterServicePoint(0, &types.HealthPoint{}, health)
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
	return ENDPOINT
}
func (this *HealthServicePoint) Topic() string {
	return TOPIC
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
