package health

import (
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"reflect"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
	typ          *reflect.Type
}

func RegisterHealth(resources common.IResources, listener cache.ICacheListener) {
	health := &HealthServicePoint{}
	health.healthCenter = newHealthCenter(resources, listener)
	err := resources.ServicePoints().RegisterServicePoint(health, 0)
	if err != nil {
		panic(err)
	}
}

func (this *HealthServicePoint) Post(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil
}
func (this *HealthServicePoint) Put(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp)
	return nil
}
func (this *HealthServicePoint) Patch(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Update(hp)
	return nil
}
func (this *HealthServicePoint) Delete(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	return nil
}
func (this *HealthServicePoint) GetCopy(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	return nil
}
func (this *HealthServicePoint) Get(pb common.IMObjects, resourcs common.IResources) common.IMObjects {
	return nil
}
func (this *HealthServicePoint) Failed(pb common.IMObjects, resourcs common.IResources, msg *types.Message) common.IMObjects {
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

func (this *HealthServicePoint) ServiceModel() common.IMObjects {
	return object.New("", &types.HealthPoint{})
}
