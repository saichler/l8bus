package health

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
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
	hc.healthPoints = cache.NewModelCache(resources.Config().LocalUuid, listener, resources.Introspector())
	hc.services = newServices()
	hc.resources = resources
	return hc
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint)
	this.services.Update(healthPoint)
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint) {
	err := this.healthPoints.Update(healthPoint.AUuid, healthPoint)
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

func (this *HealthCenter) UuidsForRequest(cast types.CastMode, vlanId int32, topic, source string) string {
	if len(topic) == protocol.UNICAST_ADDRESS_SIZE {
		return topic
	}
	uuids := this.services.UUIDs(topic, vlanId, false)
	switch cast {
	case types.CastMode_All:
		fallthrough
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
		return this.services.Leader(topic, vlanId)
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

func (this *HealthCenter) Leader(topic string, vlanId int32) string {
	return this.services.Leader(topic, vlanId)
}

func (this *HealthCenter) AllTopics() *types.Topics {
	return this.services.AllTopics()
}

func (this *HealthCenter) Uuids(topic string, vlan int32, noVnet bool) map[string]bool {
	return this.services.UUIDs(topic, vlan, noVnet)
}

func (this *HealthCenter) ReplicasFor(topicId string, vlanId int32, numOfReplicas int) map[string]int32 {
	return this.services.ReplicasFor(topicId, vlanId, numOfReplicas)
}

func (this *HealthCenter) AddScore(target, topic string, vlanId int32) {
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
	vlan, ok := hp.Topics.TopicToVlan[topic]
	if !ok {
		panic("TopicToVlan is nil!")
	}
	vlan.Vlans[vlanId]++
	this.healthPoints.Update(hp.AUuid, hp)
}

func Health(resource common.IResources) *HealthCenter {
	sp, ok := resource.ServicePoints().ServicePointHandler(TOPIC)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
