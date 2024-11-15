package protocol

import (
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/maps"
	"reflect"
)

type EdgeMap struct {
	impl *maps.SyncMap
}

var localEdge interfaces.IEdge
var edgeType = reflect.TypeOf(localEdge)

func NewEdgeMap() *EdgeMap {
	m := &EdgeMap{}
	m.impl = maps.NewSyncMap()
	return m
}

func (pm *EdgeMap) Put(key string, value interfaces.IEdge) bool {
	return pm.impl.Put(key, value)
}

func (pm *EdgeMap) Get(key string) (interfaces.IEdge, bool) {
	value, ok := pm.impl.Get(key)
	if value != nil {
		return value.(interfaces.IEdge), ok
	}
	return nil, ok
}

func (pm *EdgeMap) Contains(key string) bool {
	return pm.impl.Contains(key)
}

func (pm *EdgeMap) EdgeList() []interfaces.IEdge {
	return pm.impl.ValuesAsList(edgeType, nil).([]interfaces.IEdge)
}

func (pm *EdgeMap) Iterate(do func(k, v interface{})) {
	pm.impl.Iterate(do)
}
