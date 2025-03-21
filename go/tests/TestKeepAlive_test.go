package tests

import (
	"github.com/saichler/layer8/go/overlay/health"
	. "github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
	"time"
)

func TestKeepAlive(t *testing.T) {
	pb := &testtypes.TestProto{}
	err := eg1.Multicast(ServiceName, 0, types.Action_POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	time.Sleep(time.Second * time.Duration(eg1.Resources().Config().KeepAliveIntervalSeconds+1))
	hc := health.Health(eg3.Resources())
	hp := hc.HealthPoint(eg1.Resources().Config().LocalUuid)
	if hp.Stats.TxMsgCount == 0 {
		Log.Fail(t, "Expected at least one message to be sent for ", eg1.Resources().Config().LocalUuid)
	}
}
