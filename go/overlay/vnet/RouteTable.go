package vnet

import "sync"

type RouteTable struct {
	routes   *sync.Map
	vnetUuid string
}

func newRouteTable(vnetUuid string) *RouteTable {
	return &RouteTable{routes: &sync.Map{}, vnetUuid: vnetUuid}
}

func (this *RouteTable) addRoutes(routes map[string]string) map[string]string {
	added := make(map[string]string)
	for k, v := range routes {
		_, ok := this.routes.Load(k)
		if !ok {
			this.routes.Store(k, v)
			added[k] = v
		}
	}
	return added
}

func (this *RouteTable) vnetOf(uuid string) (string, bool) {
	vnetUuid, ok := this.routes.Load(uuid)
	if ok {
		return vnetUuid.(string), ok
	}
	return this.vnetUuid, false
}
