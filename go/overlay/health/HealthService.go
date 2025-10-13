package health

/*
import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/web"
)

type HealthService struct {
	healthCenter *HealthCenter
}

func (this *HealthService) Activate(serviceName string, serviceArea byte,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	_, err := resources.Registry().Register(&l8health.L8Health{})
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
	hp, ok := pb.Element().(*l8health.L8Health)
	if !ok {
		return nil
	}
	this.healthCenter.Put(hp, pb.Notification())
	return nil
}
func (this *HealthService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp, ok := pb.Element().(*l8health.L8Health)
	if !ok {
		return nil
	}
	this.healthCenter.Put(hp, pb.Notification())
	this.healthCenter.healths.Sync()
	return nil
}
func (this *HealthService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp := pb.Element().(*l8health.L8Health)
	this.healthCenter.Patch(hp, pb.Notification())
	return nil
}
func (this *HealthService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	hp := pb.Element().(*l8health.L8Health)
	this.healthCenter.Delete(hp, pb.Notification())
	return nil
}
func (this *HealthService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *HealthService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return object.New(nil, this.healthCenter.Top())
}
func (this *HealthService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

func (this *HealthService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

func (this *HealthService) WebService() ifs.IWebService {
	return web.New(ServiceName, 0, nil, nil, nil, nil, nil, nil, nil, nil,
		&l8web.L8Empty{}, &l8health.L8Top{})
}
*/
