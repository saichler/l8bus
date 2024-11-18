package state

import (
	"github.com/saichler/layer8/go/types"
	"github.com/saichler/shared/go/share/interfaces"
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

func NewStateCenter() *StateCenter {
	stc := &StateCenter{}
	stc.mtx = &sync.RWMutex{}
	stc.cond = sync.NewCond(stc.mtx)
	stc.edges = make(map[string]*types.EdgeInfo)
	stc.services = make(map[string]*types.ServiceInfo)

	return stc
}

func (stc *StateCenter) AddEdge(newEdge interfaces.IEdge) {
	stc.mtx.Lock()
	defer stc.mtx.Unlock()
	config := newEdge.Config()
	_, ok := stc.edges[config.Uuid]
	if !ok {
		edgeInfo := &types.EdgeInfo{}
		edgeInfo.Uuid = config.Uuid
		edgeInfo.UpSince = time.Now().Unix()
		stc.edges[config.Uuid] = edgeInfo

		serviceInfo := &types.ServiceInfo{}
		serviceInfo.Uuids = make(map[string]bool)
		serviceInfo.Uuids[config.Uuid] = true

		edgeInfo.Services = make(map[string]*types.ServiceInfo)
		edgeInfo.Services[config.Uuid] = serviceInfo

		stateService, ok := stc.services[STATE_TOPIC]
		if !ok {
			stateService = &types.ServiceInfo{}
			stateService.Uuids = make(map[string]bool)
			stc.services[STATE_TOPIC] = stateService
		}
		stateService.Uuids[config.Uuid] = true
	}
}

func (stc *StateCenter) ServiceUuids(destination, source string) []string {
	stc.mtx.RLock()
	defer stc.mtx.RUnlock()
	service, ok := stc.services[destination]
	if !ok {
		return emptyServices
	}
	result := make([]string, len(service.Uuids))
	i := 0
	for uuid, _ := range service.Uuids {
		if uuid != source {
			result[i] = uuid
		}
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
