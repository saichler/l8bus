package tests

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/shared/go/types"
	"testing"
	"time"
)

func TestKeepAlive(t *testing.T) {
	pb := &tests.TestProto{}
	err := eg1.Multicast(types.CastMode_All, types.Action_POST, 0, infra.TEST_TOPIC, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}

	time.Sleep(time.Second * time.Duration(eg1.Resources().Config().KeepAliveIntervalSeconds+1))
	hc := health.Health(eg3.Resources())
	hp := hc.HealthPoint(eg1.Resources().Config().LocalUuid)
	if hp.Stats.TxMsgCount == 0 {
		log.Fail(t, "Expected at least one message to be sent for ", eg1.Resources().Config().LocalUuid)
	}
}
