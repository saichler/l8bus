package health

import (
	"github.com/saichler/reflect/go/reflect/introspecting"
	"github.com/saichler/servicepoints/go/points/dcache"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

type HealthCenter struct {
	healthPoints ifs.IDistributedCache
	services     *Services
	resources    ifs.IResources
}

func newHealthCenter(resources ifs.IResources, listener ifs.IServicePointCacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.HealthPoint{})
	introspecting.AddPrimaryKeyDecorator(rnode, "AUuid")
	hc.healthPoints = dcache.NewDistributedCache(ServiceName, 0, "HealthPoint",
		resources.SysConfig().LocalUuid, listener, resources)
	hc.services = newServices()
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint, isNotification bool) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint, isNotification)
	this.services.Update(healthPoint)
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint, isNotification bool) {
	_, err := this.healthPoints.Update(healthPoint.AUuid, healthPoint, isNotification)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	updatedHealthPoint := this.HealthPoint(healthPoint.AUuid)
	this.services.Update(updatedHealthPoint)
}

func (this *HealthCenter) ZSide(uuid string) string {
	hp, ok := this.healthPoints.Get(uuid).(*types.HealthPoint)
	if ok {
		return hp.ZUuid
	}
	return ""
}

func (this *HealthCenter) HealthPoint(uuid string) *types.HealthPoint {
	hp, _ := this.healthPoints.Get(uuid).(*types.HealthPoint)
	return hp
}

func (this *HealthCenter) DestinationFor(serviceName string, serviceArea uint16, source string, all, leader bool) string {
	if all {
		return ""
	}
	if leader {
		return this.services.Leader(serviceName, serviceArea)
	}
	uuids := this.services.UUIDs(serviceName, serviceArea)
	_, ok := uuids[source]
	if ok {
		return source
	}
	sourceZSide := this.services.ZUuid(source)
	for uuid, _ := range uuids {
		uuidZside := this.services.ZUuid(uuid)
		if sourceZSide == uuidZside {
			return uuid
		}
	}
	return this.services.Leader(serviceName, serviceArea)
}

func healthPoint(item interface{}) (bool, interface{}) {
	hp := item.(*types.HealthPoint)
	return true, hp
}

func (this *HealthCenter) All() map[string]*types.HealthPoint {
	uuids := this.healthPoints.Collect(healthPoint)
	result := make(map[string]*types.HealthPoint)
	for k, v := range uuids {
		result[k] = v.(*types.HealthPoint)
	}
	return result
}

func (this *HealthCenter) Leader(multicast string, serviceArea uint16) string {
	return this.services.Leader(multicast, serviceArea)
}

func (this *HealthCenter) AllServices() *types.Services {
	return this.services.AllServices()
}

func (this *HealthCenter) Uuids(serviceName string, serviceArea uint16) map[string]bool {
	return this.services.UUIDs(serviceName, serviceArea)
}

func (this *HealthCenter) Top() *types.Top {
	all := this.All()
	top := &types.Top{HealthPoints: make(map[string]*types.HealthPoint)}
	for k, v := range all {
		top.HealthPoints[k] = v
	}
	return top
}

func Health(r ifs.IResources) *HealthCenter {
	sp, ok := r.ServicePoints().ServicePointHandler(ServiceName, 0)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
