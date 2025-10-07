package tests

import (
	"testing"

	"github.com/saichler/l8types/go/ifs"
)

func getLeader(uuid string) ifs.IVNic {
	all := topo.AllVnics()
	for _, nic := range all {
		if nic.Resources().SysConfig().LocalUuid == uuid {
			return nic
		}
	}
	panic("No Leader")
}

func TestLeader(t *testing.T) {
	/*
		eg2_3 := topo.VnicByVnetNum(2, 3)
		hc := health.Health(eg2_3.Resources())
		leaderBefore := hc.LeaderFor(ServiceName, 0)
		leader := getLeader(leaderBefore)
		leader.Shutdown()
		defer func() {
			topo.RenewVnic(leader.Resources().SysConfig().LocalAlias)
		}()
		time.Sleep(time.Second * 10)
		leaderAfter := hc.LeaderFor(ServiceName, 0)
		if leaderAfter == leaderBefore {
			Log.Fail(t, "Expected leader to change")
			return
		}*/
}
