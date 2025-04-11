package vnic

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"runtime"
	"time"
)

type KeepAlive struct {
	vnic *VirtualNetworkInterface
}

func newKeepAlive(vnic *VirtualNetworkInterface) *KeepAlive {
	return &KeepAlive{vnic: vnic}
}

func (this *KeepAlive) start() {
	go this.run()
}
func (this *KeepAlive) shutdown() {}
func (this *KeepAlive) name() string {
	return "KA"
}
func (this *KeepAlive) run() {
	if this.vnic.resources.SysConfig().KeepAliveIntervalSeconds == 0 {
		return
	}
	for this.vnic.running {
		for i := 0; i < int(this.vnic.resources.SysConfig().KeepAliveIntervalSeconds*10); i++ {
			time.Sleep(time.Millisecond * 100)
			if !this.vnic.running {
				return
			}
		}
		this.sendState()
	}
}

func (this *KeepAlive) sendState() {
	return
	stats := &types.HealthPointStats{}
	stats.TxMsgCount = this.vnic.stats.TxMsgCount
	stats.TxDataCount = this.vnic.stats.TxDataCount
	stats.RxMsgCount = this.vnic.stats.RxMsgCount
	stats.RxDataCont = this.vnic.stats.RxDataCont
	stats.LastMsgTime = time.Now().UnixMilli()
	stats.MemoryUsage = memoryUsage()
	stats.CpuUsage = cpuUsage()

	hp := &types.HealthPoint{}
	hp.AUuid = this.vnic.resources.SysConfig().LocalUuid
	hp.Status = types.HealthState_Up
	hp.Stats = stats
	this.vnic.resources.Logger().Debug("Sending Keep Alive for ", this.vnic.resources.SysConfig().LocalUuid, " ", this.vnic.resources.SysConfig().LocalAlias)
	//We unicast to the vnet, it will multicast the change to all
	this.vnic.Unicast(this.vnic.resources.SysConfig().RemoteUuid, health.ServiceName, 0, common.PATCH, hp)
}

func memoryUsage() uint64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats.Alloc
}

func cpuUsage() float64 {
	//pprof.StartCPUProfile()
	//@TODO implement a second profile
	return 0
}
