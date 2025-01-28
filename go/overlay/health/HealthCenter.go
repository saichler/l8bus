package health

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
	"sync"
)

type HealthCenter struct {
	mtx       *sync.RWMutex
	statuses  map[string]*types.HealthPoint
	services  map[string]map[string]bool
	resources interfaces.IResources
}

func newHealthCenter(resources interfaces.IResources) *HealthCenter {
	hc := &HealthCenter{}
	hc.statuses = make(map[string]*types.HealthPoint)
	hc.services = make(map[string]map[string]bool)
	hc.mtx = &sync.RWMutex{}
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.statuses[healthPoint.AUuid] = healthPoint
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
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.statuses[healthPoint.AUuid] = healthPoint
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
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	st, ok := this.statuses[uuid]
	if ok {
		return st.ZUuid
	}
	return ""
}

func (this *HealthCenter) GetState(uuid string) *types.HealthPoint {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.statuses[uuid]
}

func (this *HealthCenter) SetState(uuid string, state types.State) (*types.HealthPoint, bool) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	st, ok := this.statuses[uuid]
	if ok && st.Status != state {
		st.Status = state
		return st, true
	}
	return st, false
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

func (this *HealthCenter) AllPoints() map[string]*types.HealthPoint {
	result := make(map[string]*types.HealthPoint)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for k, v := range this.statuses {
		result[k] = v
	}
	return result
}

func (this *HealthCenter) Print() {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	this.resources.Logger().Info("** HealthCenter ", this.resources.Config().LocalAlias)
	for _, hp := range this.statuses {
		this.resources.Logger().Info("** -- ", hp.Alias)
		for svc, _ := range hp.Services {
			this.resources.Logger().Info("** ---- ", svc)
		}
	}
}

func Health(resource interfaces.IResources) *HealthCenter {
	sp, ok := resource.ServicePoints().ServicePointHandler(TOPIC)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
