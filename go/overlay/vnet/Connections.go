package vnet

import (
	"github.com/saichler/l8types/go/ifs"
	"sync"
)

type Connections struct {
	internal map[string]ifs.IVNic
	external map[string]ifs.IVNic
	routes   map[string]string
	mtx      *sync.RWMutex
	logger   ifs.ILogger
	vnetUuid string
}

func newConnections(vnetUuid string, logger ifs.ILogger) *Connections {
	conns := &Connections{}
	conns.internal = make(map[string]ifs.IVNic)
	conns.external = make(map[string]ifs.IVNic)
	conns.routes = make(map[string]string)
	conns.mtx = &sync.RWMutex{}
	conns.logger = logger
	conns.vnetUuid = vnetUuid
	return conns
}

func (this *Connections) addInternal(uuid string, vnic ifs.IVNic) {
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

func (this *Connections) addExternal(uuid string, vnic ifs.IVNic) {
	this.logger.Info("Adding external with alias ", vnic.Resources().SysConfig().RemoteAlias)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	exist, ok := this.external[uuid]
	if ok {
		this.logger.Info("External vnic ", uuid, " already exists, shutting down")
		exist.Shutdown()
	}
	this.external[uuid] = vnic
}

func (this *Connections) isConnected(ip string) bool {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for _, conn := range this.external {
		addr := conn.Resources().SysConfig().Address
		if ip == addr {
			return true
		}
	}
	return false
}

func (this *Connections) addRoutes(routes map[string]string) map[string]string {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	added := make(map[string]string)
	for k, v := range routes {
		_, ok := this.routes[k]
		if !ok {
			this.routes[k] = v
			added[k] = v
		}
	}
	return added
}

func (this *Connections) getConnection(vnicUuid string, isHope0 bool) (string, ifs.IVNic) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	vnic, ok := this.internal[vnicUuid]
	if ok {
		return vnicUuid, vnic
	}
	// only if this is hope0, e.g. the source of the message is from this switch sources,
	// fetch try to find the route
	if isHope0 {
		vnic, ok = this.external[vnicUuid]
		if ok {
			return vnicUuid, vnic
		}

		remoteUuid := this.routes[vnicUuid]
		vnic, ok = this.external[remoteUuid]
		if ok {
			return remoteUuid, vnic
		}
	}
	return "", nil
}

func (this *Connections) all() map[string]ifs.IVNic {
	all := make(map[string]ifs.IVNic)
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

func (this *Connections) isInterval(uuid string) bool {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	_, ok := this.internal[uuid]
	return ok
}

func (this *Connections) allInternals() map[string]ifs.IVNic {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := make(map[string]ifs.IVNic)
	for uuid, vnic := range this.internal {
		result[uuid] = vnic
	}
	return result
}

func (this *Connections) allExternals() map[string]ifs.IVNic {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	result := make(map[string]ifs.IVNic)
	for uuid, vnic := range this.external {
		result[uuid] = vnic
	}
	return result
}

func (this *Connections) shutdownConnection(uuid string) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	this.logger.Info("Shutting down connection ", uuid)
	conn, ok := this.internal[uuid]
	if ok {
		conn.Shutdown()
	}
	conn, ok = this.external[uuid]
	if ok {
		conn.Shutdown()
	}
}

func (this *Connections) allDownConnections() map[string]bool {
	result := make(map[string]bool)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for uuid, conn := range this.internal {
		if !conn.Running() {
			result[uuid] = true
		}
	}
	for uuid, conn := range this.external {
		if !conn.Running() {
			result[uuid] = true
		}
	}
	return result
}

func (this *Connections) Routes() map[string]string {
	routes := make(map[string]string)
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	for k, _ := range this.internal {
		routes[k] = this.vnetUuid
	}
	/*
		for k, v := range this.routes {
			routes[k] = v
		}*/
	return routes
}

func (this *Connections) RouteTableSize() int {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	sum := make(map[string]bool)
	/*
		for k, _ := range this.internal {
			sum[k] = true
		}
		for k, _ := range this.external {
			sum[k] = true
		}*/
	for k, _ := range this.routes {
		sum[k] = true
	}
	return len(sum)
}
