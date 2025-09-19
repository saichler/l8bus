package vnic

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8bus/go/overlay/health"
)

type CPUTracker struct {
	lastProcCPU uint64
	lastSysCPU  uint64
	lastSample  time.Time
	mu          sync.Mutex
}

type KeepAlive struct {
	vnic       *VirtualNetworkInterface
	startTime  int64
	cpuTracker *CPUTracker
}

func newKeepAlive(vnic *VirtualNetworkInterface) *KeepAlive {
	return &KeepAlive{
		vnic:       vnic,
		cpuTracker: &CPUTracker{},
	}
}

func (this *KeepAlive) start() {
	go this.run()
}
func (this *KeepAlive) shutdown() {}
func (this *KeepAlive) name() string {
	return "KA"
}
func (this *KeepAlive) run() {
	this.startTime = time.Now().UnixMilli()
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
	stats := &l8health.L8HealthStats{}
	stats.TxMsgCount = this.vnic.healthStatistics.TxMsgCount.Load()
	stats.TxDataCount = this.vnic.healthStatistics.TxDataCount.Load()
	stats.RxMsgCount = this.vnic.healthStatistics.RxMsgCount.Load()
	stats.RxDataCont = this.vnic.healthStatistics.RxDataCont.Load()
	stats.LastMsgTime = this.vnic.healthStatistics.LastMsgTime.Load()
	stats.MemoryUsage = memoryUsage()
	stats.CpuUsage = this.cpuTracker.GetCPUUsage()

	hp := &l8health.L8Health{}
	hp.AUuid = this.vnic.resources.SysConfig().LocalUuid
	hp.Status = l8health.L8HealthState_Up
	hp.Stats = stats
	hp.StartTime = this.startTime
	//this.vnic.resources.Logger().Debug("Sending Keep Alive for ", this.vnic.resources.SysConfig().LocalUuid, " ", this.vnic.resources.SysConfig().LocalAlias)
	//We unicast to the vnet, it will multicast the change to all
	this.vnic.Unicast(this.vnic.resources.SysConfig().RemoteUuid, health.ServiceName, 0, ifs.PATCH, hp)
}

func memoryUsage() uint64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats.Alloc
}

func (c *CPUTracker) GetCPUUsage() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	// Only update CPU stats every 30 seconds to reduce syscall overhead
	if now.Sub(c.lastSample) < 30*time.Second && c.lastSample.Unix() > 0 {
		// Return cached calculation or 0 if no previous sample
		return 0
	}

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
	currentProcCPU := utime + stime

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

	var currentSysCPU uint64
	for i := 1; i < len(cpuFields); i++ {
		val, _ := strconv.ParseUint(cpuFields[i], 10, 64)
		currentSysCPU += val
	}

	var cpuPercent float64
	// Calculate differential if we have previous values
	if c.lastSample.Unix() > 0 {
		procDelta := float64(currentProcCPU - c.lastProcCPU)
		sysDelta := float64(currentSysCPU - c.lastSysCPU)

		if sysDelta > 0 {
			cpuPercent = (procDelta / sysDelta) * 100.0
		}
	}

	// Update cached values
	c.lastProcCPU = currentProcCPU
	c.lastSysCPU = currentSysCPU
	c.lastSample = now

	return cpuPercent
}
