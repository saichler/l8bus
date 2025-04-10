package health

import (
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"reflect"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
	typ          *reflect.Type
}

func Activate(resources common.IResources, listener common.IServicePointCacheListener) error {
	health := &HealthServicePoint{}
	health.healthCenter = newHealthCenter(resources, listener)
	resources.ServicePoints().RegisterServicePoint(health)
	err := resources.ServicePoints().Activate(ServiceName, 0, health, nil)
	if err != nil {
		return err
	}
	return nil
}

func (this *HealthServicePoint) Post(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil
}
func (this *HealthServicePoint) Put(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil
}
func (this *HealthServicePoint) Patch(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Update(hp)
	return nil
}
func (this *HealthServicePoint) Delete(pb common.IElements, resourcs common.IResources) common.IElements {
	return nil
}
func (this *HealthServicePoint) GetCopy(pb common.IElements, resourcs common.IResources) common.IElements {
	return nil
}
func (this *HealthServicePoint) Get(pb common.IElements, resourcs common.IResources) common.IElements {
	return nil
}
func (this *HealthServicePoint) Failed(pb common.IElements, resourcs common.IResources, msg common.IMessage) common.IElements {
	return nil
}
func (this *HealthServicePoint) EndPoint() string {
	return Endpoint
}
func (this *HealthServicePoint) ServiceName() string {
	return ServiceName
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

func (this *HealthServicePoint) ServiceModel() common.IElements {
	return object.New(nil, &types.HealthPoint{})
}
