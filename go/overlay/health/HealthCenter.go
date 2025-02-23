package health

import (
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"sync"
)

type HealthCenter struct {
	mtx          *sync.RWMutex
	healthPoints *cache.Cache
	services     *types.Areas
	resources    interfaces.IResources
}

func newHealthCenter(resources interfaces.IResources, listener cache.ICacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.HealthPoint{})
	resources.Introspector().AddDecorator(types.DecoratorType_Primary, []string{"AUuid"}, rnode)
	hc.healthPoints = cache.NewModelCache(resources.Config().LocalUuid, listener, resources.Introspector())
	hc.services = &types.Areas{}
	hc.services.AreasMap = make(map[int32]*types.Area)
	hc.mtx = &sync.RWMutex{}
	hc.resources = resources
	return hc
}

func (this *HealthCenter) updateServices(areas *types.Areas) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if areas != nil {
		for areaId, area := range areas.AreasMap {
			_, ok := this.services.AreasMap[areaId]
			if !ok {
				this.services.AreasMap[areaId] = area
				continue
			}
			for topic, addrs := range area.Topics {
				_, ok := this.services.AreasMap[areaId].Topics[topic]
				if !ok {
					this.services.AreasMap[areaId].Topics[topic] = addrs
					continue
				}
				for k, v := range addrs.Uuids {
					this.services.AreasMap[areaId].Topics[topic].Uuids[k] = v
				}
			}
		}
	}
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint)
	this.updateServices(healthPoint.ServiceAreas)
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint) {
	err := this.healthPoints.Update(healthPoint.AUuid, healthPoint)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	this.updateServices(healthPoint.ServiceAreas)
}

func (this *HealthCenter) ZSide(uuid string) string {
	hp, ok := this.healthPoints.Get(uuid).(*types.HealthPoint)
	if ok {
		return hp.ZUuid
	}
	return ""
}

func (this *HealthCenter) GetHealthPoint(uuid string) *types.HealthPoint {
	hp, _ := this.healthPoints.Get(uuid).(*types.HealthPoint)
	return hp
}

func (this *HealthCenter) UuidsForTopic(areaId int32, topic string) map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	area, ok := this.services.AreasMap[areaId]
	if !ok {
		return result
	}
	addrs, ok := area.Topics[topic]
	if !ok {
		return result
	}
	for uuid, _ := range addrs.Uuids {
		result[uuid] = true
	}
	return result
}

func healthPoint(item interface{}) interface{} {
	hp := item.(*types.HealthPoint)
	return hp
}

func (this *HealthCenter) All() map[string]*types.HealthPoint {
	uuids := this.healthPoints.Collect(healthPoint)
	result := make(map[string]*types.HealthPoint)
	for k, v := range uuids {
		result[k] = v.(*types.HealthPoint)
	}
	return result
}

func Health(resource interfaces.IResources) *HealthCenter {
	sp, ok := resource.ServicePoints().ServicePointHandler(TOPIC)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
