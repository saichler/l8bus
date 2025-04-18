package health

import (
	"github.com/saichler/types/go/types"
	"sort"
	"sync"
)

type Services struct {
	services    map[string]*ServiceAreas
	aSide2zSide map[string]string
	mtx         *sync.RWMutex
}

type ServiceAreas struct {
	name  string
	areas map[uint16]*ServiceArea
}

type ServiceArea struct {
	members map[string]*Member
	leader  string
}

type Member struct {
	t int64
	s int32
}

func newServices() *Services {
	services := &Services{}
	services.services = make(map[string]*ServiceAreas)
	services.aSide2zSide = make(map[string]string)
	services.mtx = new(sync.RWMutex)
	return services
}

func (this *Services) ZUuid(auuid string) string {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.aSide2zSide[auuid]
}

func (this *Services) UUIDs(serviceName string, serviceArea uint16) map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	serviceAreas, ok := this.services[serviceName]
	if !ok {
		return result
	}
	area, ok := serviceAreas.areas[serviceArea]
	if !ok {
		return result
	}
	for uuid, _ := range area.members {
		if uuid == area.leader {
			result[uuid] = true
		} else {
			result[uuid] = false
		}
	}
	return result
}

func (this *Services) Leader(serviceName string, serviceArea uint16) string {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	serviceAreas, ok := this.services[serviceName]
	if !ok {
		return ""
	}
	area, ok := serviceAreas.areas[serviceArea]
	if !ok {
		return ""
	}
	return area.leader
}

func (this *Services) ReplicasFor(serviceName string, serviceArea uint16, numOfReplicas int) map[string]int32 {
	scores := this.ScoresFor(serviceName, serviceArea)
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

func (this *Services) ScoresFor(serviceName string, serviceArea uint16) map[string]int32 {
	result := make(map[string]int32)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	serviceAreas, ok := this.services[serviceName]
	if !ok {
		return result
	}
	area, ok := serviceAreas.areas[serviceArea]
	if !ok {
		return result
	}
	for target, member := range area.members {
		result[target] = member.s
	}
	return result
}

func (this *Services) checkHealthPointDown(healthPoint *types.HealthPoint, areasToCalc *[]*ServiceArea) {
	if healthPoint.Status != types.HealthState_Invalid_State &&
		healthPoint.Status != types.HealthState_Up {
		for _, serviceAreas := range this.services {
			for _, area := range serviceAreas.areas {
				_, ok := area.members[healthPoint.AUuid]
				if ok {
					*areasToCalc = append(*areasToCalc, area)
					delete(area.members, healthPoint.AUuid)
				}
			}
		}
	}
}

func (this *Services) updateServices(healthPoint *types.HealthPoint, areasToCalcLeader *[]*ServiceArea) {
	if healthPoint.Services == nil {
		return
	}
	for serviceName, serviceAreas := range healthPoint.Services.ServiceToAreas {
		_, ok := this.services[serviceName]
		if !ok {
			this.services[serviceName] = &ServiceAreas{}
			this.services[serviceName].name = serviceName
			this.services[serviceName].areas = make(map[uint16]*ServiceArea)
		}
		for svArea, score := range serviceAreas.Areas {
			serviceArea := uint16(svArea)
			_, ok = this.services[serviceName].areas[serviceArea]
			if !ok {
				this.services[serviceName].areas[serviceArea] = &ServiceArea{}
				this.services[serviceName].areas[serviceArea].members = make(map[string]*Member)
			}
			if healthPoint.Status != types.HealthState_Up {
				delete(this.services[serviceName].areas[serviceArea].members, healthPoint.AUuid)
				continue
			}
			if this.services[serviceName].areas[serviceArea].members[healthPoint.AUuid] == nil {
				this.services[serviceName].areas[serviceArea].members[healthPoint.AUuid] = &Member{}
			}
			if healthPoint.StartTime != 0 {
				this.services[serviceName].areas[serviceArea].members[healthPoint.AUuid].t = healthPoint.StartTime
			}
			if this.services[serviceName].areas[serviceArea].members[healthPoint.AUuid].s < score.Score {
				this.services[serviceName].areas[serviceArea].members[healthPoint.AUuid].s = score.Score
			}
			*areasToCalcLeader = append(*areasToCalcLeader, this.services[serviceName].areas[serviceArea])
		}
	}
}

func (this *Services) Update(healthPoint *types.HealthPoint) {
	areasToCalcLeader := make([]*ServiceArea, 0)

	this.mtx.Lock()
	defer this.mtx.Unlock()

	if healthPoint.AUuid != "" && healthPoint.ZUuid != "" {
		this.aSide2zSide[healthPoint.AUuid] = healthPoint.ZUuid
	}
	this.checkHealthPointDown(healthPoint, &areasToCalcLeader)
	this.updateServices(healthPoint, &areasToCalcLeader)
	for _, vlan := range areasToCalcLeader {
		calcLeader(vlan)
	}
}

func calcLeader(serviceArea *ServiceArea) {
	var minTime int64 = -1
	serviceArea.leader = ""
	for uuid, member := range serviceArea.members {
		if minTime == -1 || member.t < minTime {
			minTime = member.t
			serviceArea.leader = uuid
		}
	}
}

func (this *Services) AllServices() *types.Services {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := &types.Services{}
	result.ServiceToAreas = make(map[string]*types.ServiceAreas)
	for name, serviceNames := range this.services {
		result.ServiceToAreas[name] = &types.ServiceAreas{}
		result.ServiceToAreas[name].Areas = make(map[int32]*types.ServiceAreaInfo)
		for svArea, _ := range serviceNames.areas {
			serviceArea := int32(svArea)
			result.ServiceToAreas[name].Areas[serviceArea] = &types.ServiceAreaInfo{}
		}
	}
	return result
}
