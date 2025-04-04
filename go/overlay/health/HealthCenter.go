package health

import (
	"github.com/saichler/reflect/go/reflect/introspecting"
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
)

const (
	ServiceName = "Health"
	Endpoint    = "health"
)

type HealthCenter struct {
	healthPoints *cache.Cache
	services     *Services
	resources    common.IResources
}

func newHealthCenter(resources common.IResources, listener cache.ICacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.HealthPoint{})
	introspecting.AddPrimaryKeyDecorator(rnode, "AUuid")
	hc.healthPoints = cache.NewModelCache(ServiceName, 0, "HealthPoint",
		resources.SysConfig().LocalUuid, listener, resources.Introspector())
	hc.services = newServices()
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint)
	this.services.Update(healthPoint)
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint) {
	_, err := this.healthPoints.Update(healthPoint.AUuid, healthPoint)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	this.services.Update(healthPoint)
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
	uuids := this.services.UUIDs(serviceName, serviceArea, false)
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

func (this *HealthCenter) Uuids(serviceName string, serviceArea uint16, noVnet bool) map[string]bool {
	return this.services.UUIDs(serviceName, serviceArea, noVnet)
}

func (this *HealthCenter) ReplicasFor(serviceName string, serviceArea uint16, numOfReplicas int) map[string]int32 {
	return this.services.ReplicasFor(serviceName, serviceArea, numOfReplicas)
}

func (this *HealthCenter) AddScore(target, serviceName string, serviceArea uint16, vnic common.IVirtualNetworkInterface) {
	hp := this.healthPoints.Get(target).(*types.HealthPoint)
	if hp == nil {
		panic("HealthPoint is nil!")
	}
	if hp.Services == nil {
		panic("Services is nil!")
	}
	if hp.Services.ServiceToAreas == nil {
		panic("ServiceToAreas is nil!")
	}
	area, ok := hp.Services.ServiceToAreas[serviceName]
	if !ok {
		panic("Area is nil!")
	}
	area.Areas[int32(serviceArea)].Score++
	n, e := this.healthPoints.Update(hp.AUuid, hp)
	if n == nil && e == nil {
		panic("Something went wrong with helth notification!")
	}
	if e != nil {
		panic(e)
	}
	e = vnic.Unicast(vnic.Resources().SysConfig().RemoteUuid, ServiceName, 0, common.Notify, n)
	if e != nil {
		panic(e)
	}
}

func Health(r common.IResources) *HealthCenter {
	sp, ok := r.ServicePoints().ServicePointHandler(ServiceName, 0)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
