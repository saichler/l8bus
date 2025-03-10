package health

import (
	"github.com/saichler/types/go/types"
	"sync"
)

type Services struct {
	topics      map[string]*Topic
	aSide2zSide map[string]string
	vnetUuid    map[string]bool
	mtx         *sync.RWMutex
}

type Topic struct {
	name  string
	vlans map[int32]*Vlan
}

type Vlan struct {
	members map[string]int64
	leader  string
}

func newServices() *Services {
	services := &Services{}
	services.topics = make(map[string]*Topic)
	services.aSide2zSide = make(map[string]string)
	services.mtx = new(sync.RWMutex)
	services.vnetUuid = make(map[string]bool)
	return services
}

func (this *Services) ZUuid(auuid string) string {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.aSide2zSide[auuid]
}

func (this *Services) UUIDs(topicId string, vlanId int32, noVnet bool) map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	topic, ok := this.topics[topicId]
	if !ok {
		return result
	}
	vlan, ok := topic.vlans[vlanId]
	if !ok {
		return result
	}
	for uuid, _ := range vlan.members {
		if noVnet {
			_, ok = this.vnetUuid[uuid]
			if ok {
				continue
			}
		}
		if uuid == vlan.leader {
			result[uuid] = true
		} else {
			result[uuid] = false
		}
	}
	return result
}

func (this *Services) Leader(topicId string, vlanId int32) string {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	topic, ok := this.topics[topicId]
	if !ok {
		return ""
	}
	vlan, ok := topic.vlans[vlanId]
	if !ok {
		return ""
	}
	return vlan.leader
}

func (this *Services) checkHealthPointDown(healthPoint *types.HealthPoint, vlansToCalcLeader *[]*Vlan) {
	if healthPoint.Status != types.HealthState_Invalid_State &&
		healthPoint.Status != types.HealthState_Up {
		for _, topic := range this.topics {
			for _, vlan := range topic.vlans {
				_, ok := vlan.members[healthPoint.AUuid]
				if ok {
					*vlansToCalcLeader = append(*vlansToCalcLeader, vlan)
					delete(vlan.members, healthPoint.AUuid)
				}
			}
		}
	}
}

func (this *Services) updateTopics(healthPoint *types.HealthPoint, vlansToCalcLeader *[]*Vlan) {
	if healthPoint.Topics == nil {
		return
	}
	if healthPoint.IsVnet {
		this.vnetUuid[healthPoint.AUuid] = true
	}
	for topic, vlans := range healthPoint.Topics.TopicToVlan {
		_, ok := this.topics[topic]
		if !ok {
			this.topics[topic] = &Topic{}
			this.topics[topic].name = topic
			this.topics[topic].vlans = make(map[int32]*Vlan)
		}
		for vlanId, _ := range vlans.Vlans {
			_, ok = this.topics[topic].vlans[vlanId]
			if !ok {
				this.topics[topic].vlans[vlanId] = &Vlan{}
				this.topics[topic].vlans[vlanId].members = make(map[string]int64)
			}
			this.topics[topic].vlans[vlanId].members[healthPoint.AUuid] = healthPoint.StartTime
			*vlansToCalcLeader = append(*vlansToCalcLeader, this.topics[topic].vlans[vlanId])
		}
	}
}

func (this *Services) Update(healthPoint *types.HealthPoint) {
	vlansToCalcLeader := make([]*Vlan, 0)
	defer func() {
		for _, vlan := range vlansToCalcLeader {
			calcLeader(vlan)
		}
	}()

	this.mtx.Lock()
	defer this.mtx.Unlock()

	if healthPoint.AUuid != "" && healthPoint.ZUuid != "" {
		this.aSide2zSide[healthPoint.AUuid] = healthPoint.ZUuid
	}
	this.checkHealthPointDown(healthPoint, &vlansToCalcLeader)
	this.updateTopics(healthPoint, &vlansToCalcLeader)
}

func calcLeader(vlan *Vlan) {
	var minTime int64 = -1
	vlan.leader = ""
	for uuid, t := range vlan.members {
		if minTime == -1 || t < minTime {
			minTime = t
			vlan.leader = uuid
		}
	}
}

func (this *Services) setVnetUuid(uuid string) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.vnetUuid[uuid] = true
}

func (this *Services) AllTopics() *types.Topics {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := &types.Topics{}
	result.TopicToVlan = make(map[string]*types.Vlans)
	for name, topics := range this.topics {
		result.TopicToVlan[name] = &types.Vlans{}
		result.TopicToVlan[name].Vlans = make(map[int32]bool)
		for vlanId, _ := range topics.vlans {
			result.TopicToVlan[name].Vlans[vlanId] = true
		}
	}
	return result
}
