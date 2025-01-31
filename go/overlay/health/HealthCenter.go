package health

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/shared/go/share/interfaces"
	"sync"
)

type HealthCenter struct {
	mtx          *sync.RWMutex
	healthPoints *cache.Cache
	services     map[string]map[string]bool
	resources    interfaces.IResources
}

func newHealthCenter(resources interfaces.IResources) *HealthCenter {
	hc := &HealthCenter{}
	resources.Introspector().Inspect(&types.HealthPoint{})
	hc.healthPoints = cache.NewModelCache(resources.Config().LocalUuid, nil, resources.Introspector())
	hc.services = make(map[string]map[string]bool)
	hc.mtx = &sync.RWMutex{}
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if healthPoint.Services != nil && len(healthPoint.Services) > 0 {
		for topic, _ := range healthPoint.Services {
			uuids, ok := this.services[topic]
			if !ok {
				uuids = make(map[string]bool)
				this.services[topic] = uuids
			}
			uuids[healthPoint.AUuid] = true
		}
	}
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint) {
	err := this.healthPoints.Update(healthPoint.AUuid, healthPoint)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if healthPoint.Services != nil && len(healthPoint.Services) > 0 {
		for topic, _ := range healthPoint.Services {
			uuids, ok := this.services[topic]
			if !ok {
				uuids = make(map[string]bool)
				this.services[topic] = uuids
			}
			uuids[healthPoint.AUuid] = true
		}
	}
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

func (this *HealthCenter) UuidsForTopic(topic string) map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	uuids, ok := this.services[topic]
	if !ok {
		return nil
	}
	for uuid, _ := range uuids {
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
