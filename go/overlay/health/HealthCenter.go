package health

import (
	"sync"

	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8web"
)

type HealthCenter struct {
	healths       ifs.IDistributedCache
	resources     ifs.IResources
	roundRobin    map[string]map[byte]map[string]bool
	roundRobinMtx *sync.Mutex
}

func newHealthCenter(resources ifs.IResources, listener ifs.IServiceCacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&l8health.L8Health{})
	introspecting.AddPrimaryKeyDecorator(rnode, "AUuid")
	hc.healths = dcache.NewDistributedCache(ServiceName, 0, &l8health.L8Health{}, nil,
		listener, resources)
	hc.resources = resources
	hc.roundRobin = make(map[string]map[byte]map[string]bool)
	hc.roundRobinMtx = &sync.Mutex{}
	resources.Registry().Register(&l8web.L8Empty{})
	return hc
}

func (this *HealthCenter) Put(health *l8health.L8Health, isNotification bool) {
	this.healths.Put(health, isNotification)
}

func (this *HealthCenter) Delete(health *l8health.L8Health, isNotification bool) {
	this.healths.Delete(health, isNotification)
}

func (this *HealthCenter) Patch(health *l8health.L8Health, isNotification bool) {
	_, err := this.healths.Patch(health, isNotification)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
}

func (this *HealthCenter) ZSide(uuid string) string {
	filter := &l8health.L8Health{}
	filter.AUuid = uuid
	h, err := this.healths.Get(filter)
	if err != nil {
		return ""
	}
	hp, ok := h.(*l8health.L8Health)
	if ok {
		return hp.ZUuid
	}
	return ""
}

func (this *HealthCenter) Health(uuid string) *l8health.L8Health {
	filter := &l8health.L8Health{}
	filter.AUuid = uuid
	h, _ := this.healths.Get(filter)
	hp, _ := h.(*l8health.L8Health)
	return hp
}

func health(item interface{}) (bool, interface{}) {
	hp := item.(*l8health.L8Health)
	return true, hp
}

func (this *HealthCenter) All() map[string]*l8health.L8Health {
	uuids := this.healths.Collect(health)
	result := make(map[string]*l8health.L8Health)
	for k, v := range uuids {
		hp := v.(*l8health.L8Health)
		if hp.Status != l8health.L8HealthState_Down {
			result[k] = v.(*l8health.L8Health)
		}
	}
	return result
}

func (this *HealthCenter) Top() *l8health.L8Top {
	all := this.All()
	top := &l8health.L8Top{Healths: make(map[string]*l8health.L8Health)}
	for k, v := range all {
		top.Healths[k] = v
	}
	return top
}

func Health(r ifs.IResources) *HealthCenter {
	sp, ok := r.Services().ServiceHandler(ServiceName, 0)
	if !ok {
		return nil
	}
	return (sp.(*HealthService)).healthCenter
}
