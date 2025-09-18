package health

import (
	"strings"
	"sync"

	"github.com/saichler/l8types/go/ifs"
)

type Services struct {
	services    *sync.Map
	aSide2zSide *sync.Map
	logger      ifs.ILogger
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

func newServices(logger ifs.ILogger) *Services {
	services := &Services{}
	services.services = &sync.Map{}
	services.aSide2zSide = &sync.Map{}
	services.logger = logger
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

func (this *Services) updateServices(health *l8health.L8Health, areasToCalcLeader *[]*ServiceArea) {
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
			if health.Status != l8health.L8HealthState_Up {
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

func (this *Services) Update(health *l8health.L8Health) {
	areasToCalcLeader := make([]*ServiceArea, 0)
	if health == nil {
		return
	}
	if health.AUuid != "" && health.ZUuid != "" {
		this.aSide2zSide.Store(health.AUuid, health.ZUuid)
	}
	this.updateServices(health, &areasToCalcLeader)
	for _, vlan := range areasToCalcLeader {
		this.calcLeader(vlan)
	}
}

func (this *Services) Remove(uuid string) {
	this.aSide2zSide.Delete(uuid)
	this.services.Range(func(key, value interface{}) bool {
		serviceAreas := value.(*ServiceAreas)
		serviceAreas.areas.Range(func(key, value interface{}) bool {
			serviceArea := value.(*ServiceArea)
			leader := serviceArea.leader
			serviceArea.members.Delete(uuid)
			if uuid == leader {
				this.calcLeader(serviceArea)
			}
			return true
		})
		return true
	})
}

func (this *Services) calcLeader(serviceArea *ServiceArea) {
	if serviceArea == nil {
		this.logger.Error("service area is nil, disregarding")
		return
	}
	var minTime int64 = -1
	newLeader := ""
	serviceArea.members.Range(func(key, value interface{}) bool {
		uuid := key.(string)
		member := value.(*Member)
		if minTime == -1 || member.t < minTime {
			minTime = member.t
			newLeader = uuid
		} else if member.t == minTime {
			if strings.Compare(uuid, newLeader) == -1 {
				newLeader = uuid
			}
		}
		return true
	})
	serviceArea.leader = newLeader
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
