package vnic

import (
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/layer8/go/overlay/health"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	stats := &types.HealthStats{}
	stats.TxMsgCount = this.vnic.stats.TxMsgCount
	stats.TxDataCount = this.vnic.stats.TxDataCount
	stats.RxMsgCount = this.vnic.stats.RxMsgCount
	stats.RxDataCont = this.vnic.stats.RxDataCont
	stats.LastMsgTime = time.Now().UnixMilli()
	stats.MemoryUsage = memoryUsage()
	stats.CpuUsage = cpuUsage()

	hp := &types.Health{}
	hp.AUuid = this.vnic.resources.SysConfig().LocalUuid
	hp.Status = types.HealthState_Up
	hp.Stats = stats
	//this.vnic.resources.Logger().Debug("Sending Keep Alive for ", this.vnic.resources.SysConfig().LocalUuid, " ", this.vnic.resources.SysConfig().LocalAlias)
	//We unicast to the vnet, it will multicast the change to all
	this.vnic.Unicast(this.vnic.resources.SysConfig().RemoteUuid, health.ServiceName, 0, ifs.PATCH, hp)
}

func memoryUsage() uint64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats.Alloc
}

func cpuUsage() float64 {
	procStatData, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0
	}
	
	statFields := strings.Fields(string(procStatData))
	if len(statFields) < 17 {
		return 0
	}
	
	utime, _ := strconv.ParseUint(statFields[13], 10, 64)
	stime, _ := strconv.ParseUint(statFields[14], 10, 64)
	
	systemStatData, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}
	
	systemStatLines := strings.Split(string(systemStatData), "\n")
	cpuLine := systemStatLines[0]
	cpuFields := strings.Fields(cpuLine)
	if len(cpuFields) < 8 {
		return 0
	}
	
	var totalCPU uint64
	for i := 1; i < len(cpuFields); i++ {
		val, _ := strconv.ParseUint(cpuFields[i], 10, 64)
		totalCPU += val
	}
	
	processCPU := float64(utime + stime)
	totalCPUFloat := float64(totalCPU)
	
	if totalCPUFloat == 0 {
		return 0
	}
	
	return (processCPU / totalCPUFloat) * 100.0
}
