package health

import (
	"github.com/saichler/l8types/go/types"
	"strings"
	"sync"
)

type Services struct {
	services    map[string]*ServiceAreas
	aSide2zSide map[string]string
	mtx         *sync.RWMutex
}

type ServiceAreas struct {
	name  string
	areas map[byte]*ServiceArea
}

type ServiceArea struct {
	members map[string]*Member
	leader  string
}

type Member struct {
	t int64
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

func (this *Services) UUIDs(serviceName string, serviceArea byte) map[string]bool {
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

func (this *Services) Leader(serviceName string, serviceArea byte) string {
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

func (this *Services) checkHealthDown(health *types.Health, areasToCalc *[]*ServiceArea) {
	if health.Status != types.HealthState_Invalid_State &&
		health.Status != types.HealthState_Up {
		for _, serviceAreas := range this.services {
			for _, area := range serviceAreas.areas {
				_, ok := area.members[health.AUuid]
				if ok {
					*areasToCalc = append(*areasToCalc, area)
					delete(area.members, health.AUuid)
				}
			}
		}
	}
}

func (this *Services) updateServices(health *types.Health, areasToCalcLeader *[]*ServiceArea) {
	if health.Services == nil {
		return
	}
	for serviceName, serviceAreas := range health.Services.ServiceToAreas {
		_, ok := this.services[serviceName]
		if !ok {
			this.services[serviceName] = &ServiceAreas{}
			this.services[serviceName].name = serviceName
			this.services[serviceName].areas = make(map[byte]*ServiceArea)
		}
		for svArea, _ := range serviceAreas.Areas {
			serviceArea := byte(svArea)
			_, ok = this.services[serviceName].areas[serviceArea]
			if !ok {
				this.services[serviceName].areas[serviceArea] = &ServiceArea{}
				this.services[serviceName].areas[serviceArea].members = make(map[string]*Member)
			}
			if health.Status != types.HealthState_Up {
				delete(this.services[serviceName].areas[serviceArea].members, health.AUuid)
				continue
			}
			if this.services[serviceName].areas[serviceArea].members[health.AUuid] == nil {
				this.services[serviceName].areas[serviceArea].members[health.AUuid] = &Member{}
			}
			if health.StartTime != 0 {
				this.services[serviceName].areas[serviceArea].members[health.AUuid].t = health.StartTime
			}
			*areasToCalcLeader = append(*areasToCalcLeader, this.services[serviceName].areas[serviceArea])
		}
	}
}

func (this *Services) Update(health *types.Health) {
	areasToCalcLeader := make([]*ServiceArea, 0)

	this.mtx.Lock()
	defer this.mtx.Unlock()

	if health.AUuid != "" && health.ZUuid != "" {
		this.aSide2zSide[health.AUuid] = health.ZUuid
	}
	this.checkHealthDown(health, &areasToCalcLeader)
	this.updateServices(health, &areasToCalcLeader)
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
		} else if member.t == minTime {
			if strings.Compare(uuid, serviceArea.leader) == -1 {
				serviceArea.leader = uuid
			}
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
		result.ServiceToAreas[name].Areas = make(map[int32]bool)
		for svArea, _ := range serviceNames.areas {
			serviceArea := int32(svArea)
			result.ServiceToAreas[name].Areas[serviceArea] = true
		}
	}
	return result
}
