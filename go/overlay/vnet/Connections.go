package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/shared/go/share/interfaces"
	"sync"
)

type Connections struct {
	internal map[string]interfaces.IVirtualNetworkInterface
	external map[string]interfaces.IVirtualNetworkInterface
	routes   map[string]string
	mtx      *sync.RWMutex
	logger   interfaces.ILogger
}

func newConnections(logger interfaces.ILogger) *Connections {
	conns := &Connections{}
	conns.internal = make(map[string]interfaces.IVirtualNetworkInterface)
	conns.external = make(map[string]interfaces.IVirtualNetworkInterface)
	conns.routes = make(map[string]string)
	conns.mtx = &sync.RWMutex{}
	conns.logger = logger
	return conns
}

func (this *Connections) addInternal(uuid string, vnic interfaces.IVirtualNetworkInterface) {
	this.logger.Info("Adding internal with alias ", vnic.Resources().Config().RemoteAlias)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	exist, ok := this.internal[uuid]
	if ok {
		this.logger.Info("Internal Connection ", uuid, " already exists, shutting down")
		exist.Shutdown()
	}
	this.internal[uuid] = vnic
}

func (this *Connections) addExternal(uuid string, vnic interfaces.IVirtualNetworkInterface) {
	this.logger.Info("Adding external with alias ", vnic.Resources().Config().RemoteAlias)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	exist, ok := this.external[uuid]
	if ok {
		this.logger.Info("External vnic ", uuid, " already exists, shutting down")
		exist.Shutdown()
	}
	this.external[uuid] = vnic
}

func (this *Connections) getConnection(vnicUuid string, isHope0 bool, resources interfaces.IResources) (string, interfaces.IVirtualNetworkInterface) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	vnic, ok := this.internal[vnicUuid]
	if ok {
		return vnicUuid, vnic
	}
	vnic, ok = this.external[vnicUuid]
	if ok {
		return vnicUuid, vnic
	}
	// only if this is hope0, e.g. the source of the message is from this switch sources,
	// fetch try to find the route
	if isHope0 {
		remoteUuid := this.routes[vnicUuid]
		if remoteUuid == "" {
			remoteUuid = health.Health(resources).ZSide(vnicUuid)
			if remoteUuid != "" {
				this.mtx.RUnlock()
				this.mtx.Lock()
				this.routes[vnicUuid] = remoteUuid
				this.mtx.Unlock()
				this.mtx.RLock()
			}
		}

		vnic, ok = this.internal[remoteUuid]
		if ok {
			return remoteUuid, vnic
		}
		vnic, ok = this.external[remoteUuid]
		if ok {
			return remoteUuid, vnic
		}
	}
	return "", nil
}

func (this *Connections) all() map[string]interfaces.IVirtualNetworkInterface {
	all := make(map[string]interfaces.IVirtualNetworkInterface)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for uuid, vnic := range this.internal {
		all[uuid] = vnic
	}
	for uuid, vnic := range this.external {
		all[uuid] = vnic
	}
	return all
}

func (this *Connections) filterExternals(uuids map[string]int64) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for uuid, _ := range this.external {
		delete(uuids, uuid)
	}
}

func (this *Connections) isInterval(uuid string) bool {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	_, ok := this.internal[uuid]
	return ok
}
