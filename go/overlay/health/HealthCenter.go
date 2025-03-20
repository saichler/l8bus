package health

import (
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
)

const (
	Multicast = "Health"
	Endpoint  = "health"
)

type HealthCenter struct {
	healthPoints *cache.Cache
	services     *Services
	resources    common.IResources
}

func newHealthCenter(resources common.IResources, listener cache.ICacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.HealthPoint{})
	resources.Introspector().AddDecorator(types.DecoratorType_Primary, []string{"AUuid"}, rnode)
	hc.healthPoints = cache.NewModelCache(Multicast, "HealthPoint", resources.Config().LocalUuid, listener, resources.Introspector())
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

func (this *HealthCenter) DestinationFor(cast types.CastMode, vlanId int32, multicast, source string) string {
	if cast == types.CastMode_All {
		return ""
	}
	uuids := this.services.UUIDs(multicast, vlanId, false)
	switch cast {
	case types.CastMode_Single:
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
		fallthrough
	case types.CastMode_Leader:
		return this.services.Leader(multicast, vlanId)
	}
	return ""
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

func (this *HealthCenter) Leader(multicast string, vlanId int32) string {
	return this.services.Leader(multicast, vlanId)
}

func (this *HealthCenter) AllTopics() *types.Topics {
	return this.services.AllTopics()
}

func (this *HealthCenter) Uuids(multicast string, vlan int32, noVnet bool) map[string]bool {
	return this.services.UUIDs(multicast, vlan, noVnet)
}

func (this *HealthCenter) ReplicasFor(topicId string, vlanId int32, numOfReplicas int) map[string]int32 {
	return this.services.ReplicasFor(topicId, vlanId, numOfReplicas)
}

func (this *HealthCenter) AddScore(target, multicast string, vlanId int32, vnic common.IVirtualNetworkInterface) {
	hp := this.healthPoints.Get(target).(*types.HealthPoint)
	if hp == nil {
		panic("HealthPoint is nil!")
	}
	if hp.Topics == nil {
		panic("Topics is nil!")
	}
	if hp.Topics.TopicToVlan == nil {
		panic("TopicToVlan is nil!")
	}
	vlan, ok := hp.Topics.TopicToVlan[multicast]
	if !ok {
		panic("TopicToVlan is nil!")
	}
	vlan.Vlans[vlanId]++
	n, e := this.healthPoints.Update(hp.AUuid, hp)
	if n == nil && e == nil {
		panic("Something went wrong with helth notification!")
	}
	if e != nil {
		panic(e)
	}
	e = vnic.Multicast(types.CastMode_All, types.Action_Notify, vlanId, multicast, n)
	if e != nil {
		panic(e)
	}
}

func Health(r common.IResources) *HealthCenter {
	sp, ok := r.ServicePoints().ServicePointHandler(Multicast)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
