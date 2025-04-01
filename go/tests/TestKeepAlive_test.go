package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_servicepoints"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
	"time"
)

func TestKeepAlive(t *testing.T) {
	allVnics := topo.AllVnics()
	for _, nic := range allVnics {
		nic.Resources().SysConfig().KeepAliveIntervalSeconds = 2
	}

	defer func() {
		for _, nic := range allVnics {
			nic.Resources().SysConfig().KeepAliveIntervalSeconds = 30
		}
	}()

	pb := &testtypes.TestProto{}
	eg2_1 := topo.VnicByVnetNum(2, 1)
	eg1_2 := topo.VnicByVnetNum(1, 2)
	err := eg2_1.Multicast(ServiceName, 0, types.Action_POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	time.Sleep(time.Second * time.Duration(eg2_1.Resources().SysConfig().KeepAliveIntervalSeconds+2))
	hc := health.Health(eg1_2.Resources())
	hp := hc.HealthPoint(eg2_1.Resources().SysConfig().LocalUuid)
	if hp.Stats.TxMsgCount == 0 {
		Log.Fail(t, "Expected at least one message to be sent for ", eg2_1.Resources().SysConfig().LocalUuid)
	}
}
