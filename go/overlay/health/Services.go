package health

import (
	"github.com/saichler/types/go/types"
	"sort"
	"sync"
)

type Services struct {
	multicasts  map[string]*MulticastGroup
	aSide2zSide map[string]string
	vnetUuid    map[string]bool
	mtx         *sync.RWMutex
}

type MulticastGroup struct {
	name  string
	vlans map[int32]*Vlan
}

type Vlan struct {
	members map[string]*Member
	leader  string
}

type Member struct {
	t int64
	s int32
}

func newServices() *Services {
	services := &Services{}
	services.multicasts = make(map[string]*MulticastGroup)
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

func (this *Services) UUIDs(multicastId string, vlanId int32, noVnet bool) map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	multicast, ok := this.multicasts[multicastId]
	if !ok {
		return result
	}
	vlan, ok := multicast.vlans[vlanId]
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

func (this *Services) Leader(multicastId string, vlanId int32) string {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	multicast, ok := this.multicasts[multicastId]
	if !ok {
		return ""
	}
	vlan, ok := multicast.vlans[vlanId]
	if !ok {
		return ""
	}
	return vlan.leader
}

func (this *Services) ReplicasFor(multicastId string, vlanId int32, numOfReplicas int) map[string]int32 {
	scores := this.ScoresFor(multicastId, vlanId)
	if numOfReplicas > len(scores) {
		return scores
	}
	type member struct {
		target string
		score  int32
	}
	arr := make([]*member, 0)
	for target, score := range scores {
		arr = append(arr, &member{target, score})
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].score < arr[j].score
	})
	result := make(map[string]int32)
	for i := 0; i < numOfReplicas; i++ {
		result[arr[i].target] = arr[i].score
	}
	return result
}

func (this *Services) ScoresFor(multicastId string, vlanId int32) map[string]int32 {
	result := make(map[string]int32)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	multicast, ok := this.multicasts[multicastId]
	if !ok {
		return result
	}
	vlan, ok := multicast.vlans[vlanId]
	if !ok {
		return result
	}
	for target, member := range vlan.members {
		_, ok = this.vnetUuid[target]
		if ok {
			continue
		}
		result[target] = member.s
	}
	return result
}

func (this *Services) checkHealthPointDown(healthPoint *types.HealthPoint, vlansToCalcLeader *[]*Vlan) {
	if healthPoint.Status != types.HealthState_Invalid_State &&
		healthPoint.Status != types.HealthState_Up {
		for _, multicast := range this.multicasts {
			for _, vlan := range multicast.vlans {
				_, ok := vlan.members[healthPoint.AUuid]
				if ok {
					*vlansToCalcLeader = append(*vlansToCalcLeader, vlan)
					delete(vlan.members, healthPoint.AUuid)
				}
			}
		}
	}
}

func (this *Services) updateMulticastGroups(healthPoint *types.HealthPoint, vlansToCalcLeader *[]*Vlan) {
	if healthPoint.Topics == nil {
		return
	}
	if healthPoint.IsVnet {
		this.vnetUuid[healthPoint.AUuid] = true
	}
	for multicast, vlans := range healthPoint.Topics.TopicToVlan {
		_, ok := this.multicasts[multicast]
		if !ok {
			this.multicasts[multicast] = &MulticastGroup{}
			this.multicasts[multicast].name = multicast
			this.multicasts[multicast].vlans = make(map[int32]*Vlan)
		}
		for vlanId, score := range vlans.Vlans {
			_, ok = this.multicasts[multicast].vlans[vlanId]
			if !ok {
				this.multicasts[multicast].vlans[vlanId] = &Vlan{}
				this.multicasts[multicast].vlans[vlanId].members = make(map[string]*Member)
			}
			if this.multicasts[multicast].vlans[vlanId].members[healthPoint.AUuid] == nil {
				this.multicasts[multicast].vlans[vlanId].members[healthPoint.AUuid] = &Member{}
			}
			if healthPoint.StartTime != 0 {
				this.multicasts[multicast].vlans[vlanId].members[healthPoint.AUuid].t = healthPoint.StartTime
			}
			if this.multicasts[multicast].vlans[vlanId].members[healthPoint.AUuid].s < score {
				this.multicasts[multicast].vlans[vlanId].members[healthPoint.AUuid].s = score
			}
			*vlansToCalcLeader = append(*vlansToCalcLeader, this.multicasts[multicast].vlans[vlanId])
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
	this.updateMulticastGroups(healthPoint, &vlansToCalcLeader)
}

func calcLeader(vlan *Vlan) {
	var minTime int64 = -1
	vlan.leader = ""
	for uuid, member := range vlan.members {
		if minTime == -1 || member.t < minTime {
			minTime = member.t
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
	for name, multicasts := range this.multicasts {
		result.TopicToVlan[name] = &types.Vlans{}
		result.TopicToVlan[name].Vlans = make(map[int32]int32)
		for vlanId, _ := range multicasts.vlans {
			result.TopicToVlan[name].Vlans[vlanId] = 0
		}
	}
	return result
}
