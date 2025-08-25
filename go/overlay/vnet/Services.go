package vnet

import (
	"math"
	"strings"
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

type Services struct {
	services   *sync.Map
	routeTable *RouteTable
	roundrobin *sync.Map
}

func newServices(routeTable *RouteTable) *Services {
	return &Services{services: &sync.Map{}, routeTable: routeTable, roundrobin: &sync.Map{}}
}

func (this *Services) addService(data *types.ServiceData) {
	m1, ok := this.services.Load(data.ServiceName)
	if !ok {
		m1 = &sync.Map{}
		this.services.Store(data.ServiceName, m1)
	}
	area := byte(data.ServiceArea)
	m2, ok := m1.(*sync.Map).Load(area)
	if !ok {
		m2 = &sync.Map{}
		m1.(*sync.Map).Store(area, m2)
	}
	m2.(*sync.Map).Store(data.ServiceUuid, time.Now().UnixMilli())
}

func (this *Services) removeService(removed map[string]string) {
	for uuid, _ := range removed {
		this.services.Range(func(key, value interface{}) bool {
			m1 := value.(*sync.Map)
			m1.Range(func(key, value interface{}) bool {
				m2 := value.(*sync.Map)
				m2.Delete(uuid)
				return true
			})
			return true
		},
		)
	}
}

func (this *Services) serviceUuids(serviceName string, serviceArea byte) map[string]int64 {
	m1, ok := this.services.Load(serviceName)
	if !ok {
		return map[string]int64{}
	}
	m2, ok := m1.(*sync.Map).Load(serviceArea)
	if !ok {
		return map[string]int64{}
	}

	result := make(map[string]int64)
	m2.(*sync.Map).Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(int64)
		return true
	})

	return result
}

func (this *Services) serviceFor(serviceName string, serviceArea byte, source string, mode ifs.MulticastMode) string {
	m1, ok := this.services.Load(serviceName)
	if !ok {
		return ""
	}
	m2, ok := m1.(*sync.Map).Load(serviceArea)
	if !ok {
		return ""
	}
	result := ""
	switch mode {
	case ifs.M_Proximity:
		sourceVnet, _ := this.routeTable.vnetOf(source)
		m2.(*sync.Map).Range(func(key, value interface{}) bool {
			k := key.(string)
			result = k // make sure if there is a service,use it anyway even if there is no proximity
			v, _ := this.routeTable.vnetOf(k)
			if v == sourceVnet {
				result = k
				return false
			}
			return true
		})
	case ifs.M_Local:
		m2.(*sync.Map).Range(func(key, value interface{}) bool {
			k := key.(string)
			result = k // make sure if there is a service,use it anyway
			if k == source {
				result = k
				return false
			}
			return true
		})
	case ifs.M_Leader:
		minTime := int64(math.MaxInt64)
		m2.(*sync.Map).Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(int64)
			if v < minTime {
				result = k
				minTime = v
			} else if v == minTime {
				if strings.Compare(result, k) == -1 {
					result = k
				}
			}
			return true
		})
	case ifs.M_RoundRobin:
		svR, ok := this.roundrobin.Load(serviceName)
		if !ok {
			svR = &sync.Map{}
			this.roundrobin.Store(serviceName, svR)
		}
		rrS := svR.(*sync.Map)
		found := false
		m2.(*sync.Map).Range(func(key, value interface{}) bool {
			k := key.(string)
			result = k // make sure we have a result in anyway
			_, ok = rrS.Load(k)
			if !ok {
				rrS.Store(k, true)
				found = true
				return false
			}
			return true
		})
		if !found {
			rrS.Clear()
			rrS.Store(result, true)
		}
	case ifs.M_All:
		fallthrough
	default:
		m2.(*sync.Map).Range(func(key, value interface{}) bool {
			k := key.(string)
			result = k // make sure if there is a service,use it anyway
			if k != source {
				result = k
				return false
			}
			return true
		})
	}

	return result
}
