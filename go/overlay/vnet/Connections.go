package vnet

import (
	"sync"
	"sync/atomic"

	"github.com/saichler/l8types/go/ifs"
)

type Connections struct {
	internal     *sync.Map
	external     *sync.Map
	routeTable   *RouteTable
	logger       ifs.ILogger
	vnetUuid     string
	sizeInternal atomic.Int32
	sizeExternal atomic.Int32
}

func newConnections(vnetUuid string, routeTable *RouteTable, logger ifs.ILogger) *Connections {
	conns := &Connections{}
	conns.internal = &sync.Map{}
	conns.external = &sync.Map{}
	conns.routeTable = routeTable
	conns.logger = logger
	conns.vnetUuid = vnetUuid
	return conns
}

func (this *Connections) addInternal(uuid string, vnic ifs.IVNic) {
	this.logger.Info("Adding internal with alias ", vnic.Resources().SysConfig().RemoteAlias)
	exist, ok := this.internal.Load(uuid)
	if ok {
		this.logger.Info("Internal Connection ", uuid, " already exists, shutting down")
		existVnic := exist.(ifs.IVNic)
		existVnic.Shutdown()
		this.internal.Delete(uuid)
	}
	this.internal.Store(uuid, vnic)
	this.sizeInternal.Add(1)
}

func (this *Connections) addExternal(uuid string, vnic ifs.IVNic) {
	this.logger.Info("Adding external with alias ", vnic.Resources().SysConfig().RemoteAlias)
	exist, ok := this.external.Load(uuid)
	if ok {
		this.logger.Info("External vnic ", uuid, " already exists, shutting down")
		existVnic := exist.(ifs.IVNic)
		existVnic.Shutdown()
		this.external.Delete(uuid)
	}
	this.external.Store(uuid, vnic)
	this.sizeExternal.Add(1)
}

func (this *Connections) isConnected(ip string) bool {
	connected := false
	this.external.Range(func(key, value interface{}) bool {
		conn := value.(ifs.IVNic)
		addr := conn.Resources().SysConfig().Address
		if ip == addr {
			connected = true
			return false
		}
		return true
	})
	return connected
}

func (this *Connections) getConnection(vnicUuid string, isHope0 bool) (string, ifs.IVNic) {
	vnic, ok := this.internal.Load(vnicUuid)
	if ok {
		return vnicUuid, vnic.(ifs.IVNic)
	}
	// only if this is hope0, e.g. the source of the message is from this switch sources,
	// fetch try to find the route
	if isHope0 {
		vnic, ok = this.external.Load(vnicUuid)
		if ok {
			return vnicUuid, vnic.(ifs.IVNic)
		}
		remoteUuid := ""
		remoteUuid, ok = this.routeTable.vnetOf(vnicUuid)
		if !ok {
			return "", nil
		}
		vnic, ok = this.external.Load(remoteUuid)
		if ok {
			return remoteUuid, vnic.(ifs.IVNic)
		}
	}
	return "", nil
}

func (this *Connections) all() map[string]ifs.IVNic {
	all := make(map[string]ifs.IVNic)
	this.internal.Range(func(key, value interface{}) bool {
		all[key.(string)] = value.(ifs.IVNic)
		return true
	})
	this.external.Range(func(key, value interface{}) bool {
		all[key.(string)] = value.(ifs.IVNic)
		return true
	})
	return all
}

func (this *Connections) isInterval(uuid string) bool {
	_, ok := this.internal.Load(uuid)
	return ok
}

func (this *Connections) allInternals() map[string]ifs.IVNic {
	result := make(map[string]ifs.IVNic)
	this.internal.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(ifs.IVNic)
		return true
	})
	return result
}

func (this *Connections) allExternals() map[string]ifs.IVNic {
	result := make(map[string]ifs.IVNic)
	this.external.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(ifs.IVNic)
		return true
	})
	return result
}

func (this *Connections) shutdownConnection(uuid string) {
	this.logger.Info("Shutting down connection ", uuid)
	conn, ok := this.internal.Load(uuid)
	if ok {
		conn.(ifs.IVNic).Shutdown()
	}
	conn, ok = this.external.Load(uuid)
	if ok {
		conn.(ifs.IVNic).Shutdown()
	}
}

func (this *Connections) allDownConnections() map[string]bool {
	result := make(map[string]bool)
	this.internal.Range(func(key, value interface{}) bool {
		if !value.(ifs.IVNic).Running() {
			result[key.(string)] = true
		}
		return true
	})
	this.external.Range(func(key, value interface{}) bool {
		if !value.(ifs.IVNic).Running() {
			result[key.(string)] = true
		}
		return true
	})
	return result
}

func (this *Connections) Routes() map[string]string {
	routes := make(map[string]string)
	this.internal.Range(func(key, value interface{}) bool {
		routes[key.(string)] = this.vnetUuid
		return true
	})
	/*
		for k, v := range this.routes {
			routes[k] = v
		}*/
	return routes
}
