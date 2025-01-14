//go:build unit

package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/state"
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/shared/go/types"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func reset(name string) {
	interfaces.Info("*** ", name, " end ***")
	for _, t := range tsps {
		t.PostNumber = 0
		t.DeleteNumber = 0
		t.PutNumber = 0
		t.PatchNumber = 0
		t.GetNumber = 0
	}
}

func setup() {
	setupTopology()
}

func tear() {
	shutdownTopology()
}

func TestPrintTopology(t *testing.T) {
	defer reset("TestPrintTopology")
	egImpl := eg1.(*edge.EdgeImpl)
	interfaces.Info("Edge 1")
	state.Print(egImpl.State(), egImpl.Config().Local_Uuid)
	interfaces.Info("Edge 3")
	egImpl = eg3.(*edge.EdgeImpl)
	state.Print(egImpl.State(), egImpl.Config().Local_Uuid)
	interfaces.Info("Switch 1")
	state.Print(sw1.State(), sw1.Config().Local_Uuid)
}

func TestSendMultiCast(t *testing.T) {
	defer reset("TestSendMultiCast")
	pb := &tests.TestProto{}
	err := eg1.Do(types.Action_POST, infra.TEST_TOPIC, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()

	for eg, tsp := range tsps {
		if tsp.PostNumber != 1 && eg != "eg5" {
			interfaces.Fail(t, eg, " Post count does not equal 1")
			return
		} else if tsp.PostNumber != 0 && eg == "eg5" {
			interfaces.Fail(t, eg, " Post count does not equal 0")
			return
		}
	}
}

func TestUniCast(t *testing.T) {
	defer reset("TestUniCast")
	pb := &tests.TestProto{}
	err := eg2.Do(types.Action_POST, eg3.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostNumber != 1 {
		interfaces.Fail(t, "eg3", " Post count does not equal 1")
		return
	}
}

func TestReconnect(t *testing.T) {
	defer reset("TestReconnect")
	pb := &tests.TestProto{}
	err := eg5.Do(types.Action_POST, eg3.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostNumber != 1 {
		interfaces.Fail(t, "eg3", " Post count does not equal 1")
		return
	}

	interfaces.Info("********* Starting Reconnect Test")

	//Create a larger than max data
	//sending it will disconnect the socket and attempt a reconnect
	data := make([]byte, eg5.Config().MaxDataSize+1)
	eg5.Send(data)
	sleep()

	err = eg5.Do(types.Action_POST, eg3.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostNumber != 2 {
		interfaces.Fail(t, "eg3", " Post count does not equal 2 after reconnect")
		return
	}
}

func TestDestinationUnreachable(t *testing.T) {
	defer reset("TestDestinationUnreachable")
	pb := &tests.TestProto{}
	err := eg2.Do(types.Action_POST, eg5.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg5"].PostNumber != 1 {
		interfaces.Fail(t, "eg5", " Post count does not equal 1")
		return
	}

	eg5.Shutdown()
	sleep()
	err = eg2.Do(types.Action_POST, eg5.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	sleep()
}
