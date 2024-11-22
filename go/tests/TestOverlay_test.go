package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/state"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/shared/go/types"
	"testing"
	"time"
)

func TestOverlay(t *testing.T) {
	defer shutdownTopology()
	time.Sleep(time.Second * 3)
	interfaces.Info("*****************************************************************")
	time.Sleep(time.Second * 3)
	egImpl := eg1.(*edge.EdgeImpl)
	state.Print(egImpl.State(), egImpl.Config().Local_Uuid)

	egImpl = eg3.(*edge.EdgeImpl)
	state.Print(egImpl.State(), egImpl.Config().Local_Uuid)

	state.Print(sw1.State(), sw1.Config().Local_Uuid)

	pb := &tests.TestProto{}
	data, err := protocol.CreateMessageFor(types.Priority_P0, types.Action_POST, eg1.Config().Local_Uuid, eg1.Config().RemoteUuid, infra.TEST_TOPIC, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	interfaces.Info("Sending data")
	err = eg1.Send(data)
	time.Sleep(time.Second * 3)

	for eg, tsp := range tsps {
		if tsp.PostNumber != 1 {
			interfaces.Fail(t, eg, " Post count does not equal 1")
			return
		}
	}

	interfaces.Info("*****************************************************************")
}
