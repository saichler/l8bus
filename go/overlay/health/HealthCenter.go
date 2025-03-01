package health

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/servicepoints/go/points/cache"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/types"
	"sync"
)

type HealthCenter struct {
	mtx          *sync.RWMutex
	healthPoints *cache.Cache
	services     *types.Vlans
	resources    interfaces.IResources
}

func newHealthCenter(resources interfaces.IResources, listener cache.ICacheListener) *HealthCenter {
	hc := &HealthCenter{}
	rnode, _ := resources.Introspector().Inspect(&types.HealthPoint{})
	resources.Introspector().AddDecorator(types.DecoratorType_Primary, []string{"AUuid"}, rnode)
	hc.healthPoints = cache.NewModelCache(resources.Config().LocalUuid, listener, resources.Introspector())
	hc.services = &types.Vlans{}
	hc.services.Vlans = make(map[int32]*types.Vlan)
	hc.mtx = &sync.RWMutex{}
	hc.resources = resources
	return hc
}

func (this *HealthCenter) updateServices(vlans *types.Vlans) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if vlans != nil {
		for vlanId, vlan := range vlans.Vlans {
			_, ok := this.services.Vlans[vlanId]
			if !ok {
				this.services.Vlans[vlanId] = vlan
				continue
			}
			for topic, members := range vlan.Members {
				_, ok = this.services.Vlans[vlanId].Members[topic]
				if !ok {
					this.services.Vlans[vlanId].Members[topic] = members
					continue
				}
				for k, v := range members.MemberToJoinTime {
					this.services.Vlans[vlanId].Members[topic].MemberToJoinTime[k] = v
				}
			}
		}
	}
}

func (this *HealthCenter) Add(healthPoint *types.HealthPoint) {
	this.healthPoints.Put(healthPoint.AUuid, healthPoint)
	this.updateServices(healthPoint.Vlans)
}

func (this *HealthCenter) Update(healthPoint *types.HealthPoint) {
	err := this.healthPoints.Update(healthPoint.AUuid, healthPoint)
	if err != nil {
		this.resources.Logger().Error("Error updating health point ", err)
		return
	}
	this.updateServices(healthPoint.Vlans)
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

func (this *HealthCenter) UuidsForTopic(vlanId int32, topic string) map[string]int64 {
	result := make(map[string]int64)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	vlan, ok := this.services.Vlans[vlanId]
	if !ok {
		return result
	}
	members, ok := vlan.Members[topic]
	if !ok {
		return result
	}
	for uuid, joinTime := range members.MemberToJoinTime {
		result[uuid] = joinTime
	}
	return result
}

func (this *HealthCenter) UuidsForRequest(cast types.CastMode, vlanId int32, topic, source string) string {
	if len(topic) == protocol.UNICAST_ADDRESS_SIZE {
		return topic
	}
	uuids := this.UuidsForTopic(vlanId, topic)
	switch cast {
	case types.CastMode_All:
		fallthrough
	case types.CastMode_Single:
		_, ok := uuids[source]
		if ok {
			return source
		}
		sourceHp := this.healthPoints.Get(source).(*types.HealthPoint)
		leader := ""
		started := int64(-1)
		for uuid, _ := range uuids {
			uuidHp := this.healthPoints.Get(uuid).(*types.HealthPoint)
			if sourceHp.ZUuid == uuidHp.ZUuid {
				return uuid
			}
			if uuidHp.Status == types.HealthState_Up && (uuidHp.StartTime < started || started == -1) {
				leader = uuid
				started = uuidHp.StartTime
			}
		}
		return leader
	case types.CastMode_Leader:
		leader := ""
		started := int64(-1)
		for uuid, _ := range uuids {
			uuidHp := this.healthPoints.Get(uuid).(*types.HealthPoint)
			if uuidHp.Status == types.HealthState_Up && (uuidHp.StartTime < started || started == -1) {
				leader = uuid
				started = uuidHp.StartTime
			}
		}
		return leader
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

/*
func (this *HealthCenter) Leader(area int32, topic string) map[string]bool {

}*/

func Health(resource interfaces.IResources) *HealthCenter {
	sp, ok := resource.ServicePoints().ServicePointHandler(TOPIC)
	if !ok {
		return nil
	}
	return (sp.(*HealthServicePoint)).healthCenter
}
