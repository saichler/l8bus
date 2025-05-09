package health

import (
	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/reflect/go/reflect/introspecting"
)

type HealthCenter struct {
	healths   ifs.IDistributedCache
	services  *Services
	resources ifs.IResources
}

func newHealthCenter(resources ifs.IResources, listener ifs.IServiceCacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.Health{})
	introspecting.AddPrimaryKeyDecorator(rnode, "AUuid")
	hc.healths = dcache.NewDistributedCache(ServiceNames, 0, "Health",
		resources.SysConfig().LocalUuid, listener, resources)
	hc.services = newServices()
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(health *types.Health, isNotification bool) {
	this.healths.Put(health.AUuid, health, isNotification)
	this.services.Update(health)
}

func (this *HealthCenter) Update(health *types.Health, isNotification bool) {
	_, err := this.healths.Update(health.AUuid, health, isNotification)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	updatedHealth := this.Health(health.AUuid)
	this.services.Update(updatedHealth)
}

func (this *HealthCenter) ZSide(uuid string) string {
	hp, ok := this.healths.Get(uuid).(*types.Health)
	if ok {
		return hp.ZUuid
	}
	return ""
}

func (this *HealthCenter) Health(uuid string) *types.Health {
	hp, _ := this.healths.Get(uuid).(*types.Health)
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

func health(item interface{}) (bool, interface{}) {
	hp := item.(*types.Health)
	return true, hp
}

func (this *HealthCenter) All() map[string]*types.Health {
	uuids := this.healths.Collect(health)
	result := make(map[string]*types.Health)
	for k, v := range uuids {
		result[k] = v.(*types.Health)
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
	top := &types.Top{Healths: make(map[string]*types.Health)}
	for k, v := range all {
		top.Healths[k] = v
	}
	return top
}

func Health(r ifs.IResources) *HealthCenter {
	sp, ok := r.Services().ServiceHandler(ServiceNames, 0)
	if !ok {
		return nil
	}
	return (sp.(*HealthService)).healthCenter
}
