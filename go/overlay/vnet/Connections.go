package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/l8types/go/ifs"
	"sync"
)

type Connections struct {
	internal          map[string]ifs.IVirtualNetworkInterface
	external          map[string]ifs.IVirtualNetworkInterface
	routes            map[string]string
	mtx               *sync.RWMutex
	logger            ifs.ILogger
	externalConnected map[string]string
}

func newConnections(logger ifs.ILogger) *Connections {
	conns := &Connections{}
	conns.internal = make(map[string]ifs.IVirtualNetworkInterface)
	conns.external = make(map[string]ifs.IVirtualNetworkInterface)
	conns.routes = make(map[string]string)
	conns.externalConnected = make(map[string]string)
	conns.mtx = &sync.RWMutex{}
	conns.logger = logger
	return conns
}

func (this *Connections) addInternal(uuid string, vnic ifs.IVirtualNetworkInterface) {
	this.logger.Info("Adding internal with alias ", vnic.Resources().SysConfig().RemoteAlias)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	exist, ok := this.internal[uuid]
	if ok {
		this.logger.Info("Internal Connection ", uuid, " already exists, shutting down")
		exist.Shutdown()
		delete(this.internal, uuid)
	}
	this.internal[uuid] = vnic
}

func (this *Connections) addExternal(uuid string, vnic ifs.IVirtualNetworkInterface) {
	this.logger.Info("Adding external with alias ", vnic.Resources().SysConfig().RemoteAlias)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	exist, ok := this.external[uuid]
	if ok {
		this.logger.Info("External vnic ", uuid, " already exists, shutting down")
		exist.Shutdown()
	}

	this.external[uuid] = vnic
	this.externalConnected[vnic.Resources().SysConfig().Address] = uuid
}

func (this *Connections) isConnected(ip string) bool {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	_, ok := this.externalConnected[ip]
	this.logger.Info("checked ", ip, " result ", ok)
	return ok
}

func (this *Connections) getConnection(vnicUuid string, isHope0 bool, resources ifs.IResources) (string, ifs.IVirtualNetworkInterface) {
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

func (this *Connections) all() map[string]ifs.IVirtualNetworkInterface {
	all := make(map[string]ifs.IVirtualNetworkInterface)
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

func (this *Connections) filterExternals(uuids map[string]bool) {
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

func (this *Connections) allInternals() map[string]ifs.IVirtualNetworkInterface {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := make(map[string]ifs.IVirtualNetworkInterface)
	for uuid, vnic := range this.internal {
		result[uuid] = vnic
	}
	return result
}

func (this *Connections) allExternals() map[string]ifs.IVirtualNetworkInterface {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := make(map[string]ifs.IVirtualNetworkInterface)
	for uuid, vnic := range this.external {
		result[uuid] = vnic
	}
	return result
}
