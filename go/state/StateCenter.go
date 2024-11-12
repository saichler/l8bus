package state

import (
	"github.com/saichler/overlayK8s/go/types"
	"github.com/saichler/shared/go/interfaces"
	sharedTypes "github.com/saichler/shared/go/types"
	"sync"
	"time"
)

const (
	STATE_TOPIC = "STATE"
)

var emptyServices = make([]string, 0)

type StateCenter struct {
	mtx              *sync.RWMutex
	cond             *sync.Cond
	messagesSent     int64
	messagesReceived int64
	edges            map[string]*types.EdgeInfo
	services         map[string]*types.ServiceInfo
}

func (stc *StateCenter) AddEdge(newEdge interfaces.IEdge) {
	stc.mtx.Lock()
	defer stc.mtx.Unlock()
	_, ok := stc.edges[newEdge.Uuid()]
	if !ok {
		edgeInfo := &types.EdgeInfo{}
		edgeInfo.Uuid = newEdge.Uuid()
		edgeInfo.UpSince = time.Now().Unix()
		stc.edges[newEdge.Uuid()] = edgeInfo

		serviceInfo := &types.ServiceInfo{}
		serviceInfo.Uuids = make(map[string]bool)
		serviceInfo.Uuids[newEdge.Uuid()] = true

		edgeInfo.Services = make(map[string]*types.ServiceInfo)
		edgeInfo.Services[newEdge.Uuid()] = serviceInfo

		stateService := stc.services[STATE_TOPIC]
		stateService.Uuids[newEdge.Uuid()] = true
	}
}

func (stc *StateCenter) ServiceUuids(destination string) []string {
	stc.mtx.RLock()
	defer stc.mtx.RUnlock()
	service, ok := stc.services[destination]
	if !ok {
		return emptyServices
	}
	result := make([]string, len(service.Uuids))
	i := 0
	for uuid, _ := range service.Uuids {
		result[i] = uuid
		i++
	}
	return result
}

func (stc *StateCenter) InfosRequest() (*sharedTypes.Request, *types.EdgeInfos) {
	stc.mtx.RLock()
	defer stc.mtx.RUnlock()
	infoes := &types.EdgeInfos{Infos: make([]*types.EdgeInfo, len(stc.edges))}
	i := 0
	for _, edgeInfo := range stc.edges {
		infoes.Infos[i] = edgeInfo
		i++
	}
	request := &sharedTypes.Request{}
	request.Type = sharedTypes.Action_POST
	return request, infoes
}
