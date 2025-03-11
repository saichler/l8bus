package tests

import (
	"github.com/saichler/layer8/go/overlay/health"
	. "github.com/saichler/shared/go/tests/infra"
	"testing"
)

func TestLeader(t *testing.T) {
	hc := health.Health(eg3.Resources())
	leaderBefore := hc.Leader("TestProto", 0)
	eg1.Shutdown()
	defer func() {
		eg1 = createEdge(50000, "eg1", true)
		sleep()
	}()
	sleep()
	sleep()
	leaderAfter := hc.Leader("TestProto", 0)
	if leaderAfter == leaderBefore {
		Log.Fail(t, "Expected leader to change")
	}
}
