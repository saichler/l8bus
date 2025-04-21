package health

import (
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
)

const (
	ServiceName      = "Health"
	ServicePointName = "HealthServicePoint"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
}

func (this *HealthServicePoint) Activate(serviceName string, serviceArea uint16,
	resources common.IResources, listener common.IServicePointCacheListener, args ...interface{}) error {
	_, err := resources.Registry().Register(&types.HealthPoint{})
	if err != nil {
		return err
	}
	this.healthCenter = newHealthCenter(resources, listener)
	return nil
}

func (this *HealthServicePoint) DeActivate() error {
	return nil
}

func (this *HealthServicePoint) Post(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp, pb.IsNotification())
	return nil
}
func (this *HealthServicePoint) Put(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp, pb.IsNotification())
	return nil
}
func (this *HealthServicePoint) Patch(pb common.IElements, resourcs common.IResources) common.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Update(hp, pb.IsNotification())
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

func (this *HealthServicePoint) Transactional() bool {
	return false
}

func (this *HealthServicePoint) ReplicationCount() int {
	return 0
}

func (this *HealthServicePoint) ReplicationScore() int {
	return 0
}
