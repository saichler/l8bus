package vnet

import (
	"sync"
	"sync/atomic"

	"github.com/saichler/l8types/go/ifs"
)

type Connections struct {
	internal         *sync.Map
	externalVnet     *sync.Map
	externalVnic     *sync.Map
	routeTable       *RouteTable
	logger           ifs.ILogger
	vnetUuid         string
	sizeInternal     atomic.Int32
	sizeExternalVnet atomic.Int32
	sizeExternalVnic atomic.Int32
}

func newConnections(vnetUuid string, routeTable *RouteTable, logger ifs.ILogger) *Connections {
	conns := &Connections{}
	conns.internal = &sync.Map{}
	conns.externalVnet = &sync.Map{}
	conns.externalVnic = &sync.Map{}
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

func (this *Connections) addExternalVnet(uuid string, vnic ifs.IVNic) {
	this.logger.Info("Adding external with alias ", vnic.Resources().SysConfig().RemoteAlias)
	exist, ok := this.externalVnet.Load(uuid)
	if ok {
		this.logger.Info("External vnet ", uuid, " already exists, shutting down")
		existVnic := exist.(ifs.IVNic)
		existVnic.Shutdown()
		this.externalVnet.Delete(uuid)
	}
	this.externalVnet.Store(uuid, vnic)
	this.sizeExternalVnet.Add(1)
}

func (this *Connections) addExternalVnic(uuid string, vnic ifs.IVNic) {
	this.logger.Info("Adding external vnic with alias ", vnic.Resources().SysConfig().RemoteAlias)
	exist, ok := this.externalVnic.Load(uuid)
	if ok {
		this.logger.Info("External vnic ", uuid, " already exists, shutting down")
		existVnic := exist.(ifs.IVNic)
		existVnic.Shutdown()
		this.externalVnic.Delete(uuid)
	}
	this.externalVnic.Store(uuid, vnic)
	this.sizeExternalVnic.Add(1)
}

func (this *Connections) isConnected(ip string) bool {
	connected := false
	this.externalVnet.Range(func(key, value interface{}) bool {
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
	//internal vnic
	vnic, ok := this.internal.Load(vnicUuid)
	if ok {
		return vnicUuid, vnic.(ifs.IVNic)
	}
	//external vnic
	vnic, ok = this.externalVnic.Load(vnicUuid)
	if ok {
		return vnicUuid, vnic.(ifs.IVNic)
	}
	// only if this is hope0, e.g. the source of the message is from this switch sources,
	// fetch try to find the route
	if isHope0 {
		vnic, ok = this.externalVnet.Load(vnicUuid)
		if ok {
			return vnicUuid, vnic.(ifs.IVNic)
		}
		remoteUuid := ""
		remoteUuid, ok = this.routeTable.vnetOf(vnicUuid)
		if !ok {
			return "", nil
		}
		vnic, ok = this.externalVnet.Load(remoteUuid)
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
	this.externalVnet.Range(func(key, value interface{}) bool {
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

func (this *Connections) allExternalVnets() map[string]ifs.IVNic {
	result := make(map[string]ifs.IVNic)
	this.externalVnet.Range(func(key, value interface{}) bool {
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
	conn, ok = this.externalVnet.Load(uuid)
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
	this.externalVnet.Range(func(key, value interface{}) bool {
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
