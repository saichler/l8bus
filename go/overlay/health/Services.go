package health

import (
	"strings"
	"sync"

	"github.com/saichler/l8types/go/types"
)

type Services struct {
	services    *sync.Map
	aSide2zSide *sync.Map
}

type ServiceAreas struct {
	name  string
	areas *sync.Map
}

type ServiceArea struct {
	members *sync.Map
	leader  string
}

type Member struct {
	t int64
}

func newServices() *Services {
	services := &Services{}
	services.services = &sync.Map{}
	services.aSide2zSide = &sync.Map{}
	return services
}

func (this *Services) ZUuid(auuid string) string {
	zuuid, ok := this.aSide2zSide.Load(auuid)
	if ok {
		return zuuid.(string)
	}
	return ""
}

func (this *Services) UUIDs(serviceName string, serviceArea byte) map[string]bool {
	result := make(map[string]bool)
	serviceAreas, ok := this.services.Load(serviceName)
	if !ok {
		return result
	}
	area, ok := serviceAreas.(*ServiceAreas).areas.Load(serviceArea)
	if !ok {
		return result
	}

	svArea := area.(*ServiceArea)
	svArea.members.Range(func(key, value interface{}) bool {
		k := key.(string)
		if k == svArea.leader {
			result[k] = true
		} else {
			result[k] = false
		}
		return true
	})
	return result
}

func (this *Services) Leader(serviceName string, serviceArea byte) string {
	serviceAreas, ok := this.services.Load(serviceName)
	if !ok {
		return ""
	}
	area, ok := serviceAreas.(*ServiceAreas).areas.Load(serviceArea)
	if !ok {
		return ""
	}
	return area.(*ServiceArea).leader
}

func (this *Services) checkHealthDown(health *types.Health, areasToCalc *[]*ServiceArea) {
	if health.Status != types.HealthState_Invalid_State &&
		health.Status != types.HealthState_Up {
		this.services.Range(func(key, value interface{}) bool {
			value.(*ServiceAreas).areas.Range(func(key, value interface{}) bool {
				serviceArea := value.(*ServiceArea)
				_, ok := serviceArea.members.Load(health.AUuid)
				if ok {
					*areasToCalc = append(*areasToCalc, serviceArea)
					serviceArea.members.Delete(health.AUuid)
				}
				return true
			})
			return true
		})
	}
}

func (this *Services) updateServices(health *types.Health, areasToCalcLeader *[]*ServiceArea) {
	if health.Services == nil {
		return
	}
	for serviceName, serviceAreas := range health.Services.ServiceToAreas {
		existServiceAreas, ok := this.services.Load(serviceName)
		if !ok {
			newServiceAreas := &ServiceAreas{}
			newServiceAreas.name = serviceName
			newServiceAreas.areas = &sync.Map{}
			this.services.Store(serviceName, newServiceAreas)
			existServiceAreas = newServiceAreas
		}

		for svArea, _ := range serviceAreas.Areas {
			serviceArea := byte(svArea)
			existServiceArea, ok := existServiceAreas.(*ServiceAreas).areas.Load(serviceArea)
			if !ok {
				newServiceArea := &ServiceArea{}
				newServiceArea.members = &sync.Map{}
				existServiceAreas.(*ServiceAreas).areas.Store(serviceArea, newServiceArea)
				existServiceArea = newServiceArea
			}
			if health.Status != types.HealthState_Up {
				existServiceArea.(*ServiceArea).members.Delete(health.AUuid)
				continue
			}
			existMember, ok := existServiceArea.(*ServiceArea).members.Load(health.AUuid)
			if !ok {
				newMember := &Member{}
				existServiceArea.(*ServiceArea).members.Store(health.AUuid, newMember)
				existMember = newMember
			}
			if health.StartTime != 0 {
				existMember.(*Member).t = health.StartTime
			}
			*areasToCalcLeader = append(*areasToCalcLeader, existServiceArea.(*ServiceArea))
		}
	}
}

func (this *Services) Update(health *types.Health) {
	areasToCalcLeader := make([]*ServiceArea, 0)

	if health.AUuid != "" && health.ZUuid != "" {
		this.aSide2zSide.Store(health.AUuid, health.ZUuid)
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
	serviceArea.members.Range(func(key, value interface{}) bool {
		uuid := key.(string)
		member := value.(*Member)
		if minTime == -1 || member.t < minTime {
			minTime = member.t
			serviceArea.leader = uuid
		} else if member.t == minTime {
			if strings.Compare(uuid, serviceArea.leader) == -1 {
				serviceArea.leader = uuid
			}
		}
		return true
	})
}

func (this *Services) AllServices() *types.Services {
	result := &types.Services{}
	result.ServiceToAreas = make(map[string]*types.ServiceAreas)
	this.services.Range(func(key, value interface{}) bool {
		name := key.(string)
		serviceNames := value.(*ServiceAreas)
		result.ServiceToAreas[name] = &types.ServiceAreas{}
		result.ServiceToAreas[name].Areas = make(map[int32]bool)
		serviceNames.areas.Range(func(key, value interface{}) bool {
			serviceArea := int32(key.(byte))
			result.ServiceToAreas[name].Areas[serviceArea] = true
			return true
		})
		return true
	})
	return result
}
