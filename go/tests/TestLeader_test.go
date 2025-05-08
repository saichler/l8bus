package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_servicepoints"
	. "github.com/saichler/l8test/go/infra/t_topology"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/l8types/go/ifs"
	"testing"
)

func getLeader(uuid string) ifs.IVirtualNetworkInterface {
	all := topo.AllVnics()
	for _, nic := range all {
		if nic.Resources().SysConfig().LocalUuid == uuid {
			return nic
		}
	}
	panic("No Leader")
}

func testLeader(t *testing.T) {
	eg2_3 := topo.VnicByVnetNum(2, 3)
	hc := health.Health(eg2_3.Resources())
	leaderBefore := hc.Leader(ServiceName, 0)
	leader := getLeader(leaderBefore)
	leader.Shutdown()
	defer func() {
		topo.RenewVnic(leader.Resources().SysConfig().LocalAlias)
	}()
	Sleep()
	Sleep()
	leaderAfter := hc.Leader(ServiceName, 0)
	if leaderAfter == leaderBefore {
		Log.Fail(t, "Expected leader to change")
		return
	}
}
