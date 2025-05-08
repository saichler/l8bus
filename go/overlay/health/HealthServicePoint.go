package health

import (
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

const (
	ServiceName      = "Health"
	ServicePointName = "HealthServicePoint"
)

type HealthServicePoint struct {
	healthCenter *HealthCenter
}

func (this *HealthServicePoint) Activate(serviceName string, serviceArea uint16,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
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

func (this *HealthServicePoint) Post(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp, pb.Notification())
	return nil
}
func (this *HealthServicePoint) Put(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Add(hp, pb.Notification())
	this.healthCenter.healthPoints.Sync()
	return nil
}
func (this *HealthServicePoint) Patch(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	hp := pb.Element().(*types.HealthPoint)
	this.healthCenter.Update(hp, pb.Notification())
	return nil
}
func (this *HealthServicePoint) Delete(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *HealthServicePoint) GetCopy(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *HealthServicePoint) Get(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return object.New(nil, this.healthCenter.Top())
}
func (this *HealthServicePoint) Failed(pb ifs.IElements, resourcs ifs.IResources, msg ifs.IMessage) ifs.IElements {
	return nil
}

func (this *HealthServicePoint) TransactionMethod() ifs.ITransactionMethod {
	return nil
}
