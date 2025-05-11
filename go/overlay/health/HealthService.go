package health

import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

const (
	ServiceName     = "Health"
	ServiceTypeName = "HealthService"
)

type HealthService struct {
	healthCenter *HealthCenter
}

func (this *HealthService) Activate(serviceName string, serviceArea uint16,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	_, err := resources.Registry().Register(&types.Health{})
	if err != nil {
		return err
	}
	this.healthCenter = newHealthCenter(resources, listener)
	return nil
}

func (this *HealthService) DeActivate() error {
	return nil
}

func (this *HealthService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp := pb.Element().(*types.Health)
	this.healthCenter.Add(hp, pb.Notification())
	return nil
}
func (this *HealthService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp := pb.Element().(*types.Health)
	this.healthCenter.Add(hp, pb.Notification())
	this.healthCenter.healths.Sync()
	return nil
}
func (this *HealthService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp := pb.Element().(*types.Health)
	this.healthCenter.Update(hp, pb.Notification())
	return nil
}
func (this *HealthService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *HealthService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *HealthService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return object.New(nil, this.healthCenter.Top())
}
func (this *HealthService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg ifs.IMessage) ifs.IElements {
	return nil
}

func (this *HealthService) TransactionMethod() ifs.ITransactionMethod {
	return nil
}
