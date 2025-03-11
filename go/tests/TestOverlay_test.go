package tests

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/shared/go/tests/infra"
	. "github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestTopology(t *testing.T) {
	defer reset("TestPrintTopology")
	eg4Points := health.Health(eg4.Resources()).All()
	eg5Points := health.Health(eg5.Resources()).All()
	if len(eg5Points) != len(eg4Points) {
		Log.Fail(t, "Expected health points to be equal")
		return
	}
	for k, _ := range eg4Points {
		delete(eg5Points, k)
	}
	if len(eg5Points) != 0 {
		Log.Fail(t, "Expected health points to be empty")
	}

	hc := health.Health(eg2.Resources())
	uuids := hc.Uuids("TestProto", 0, false)
	if len(uuids) != 5 {
		Log.Fail(t, "Expected uuids to be 5, but it is ", len(uuids))
	}
	uuids = hc.Uuids("TestProto", 0, true)
	if len(uuids) != 4 {
		Log.Fail(t, "Expected uuids to be 4, but it is ", len(uuids))
	}
}

func TestSendMultiCast(t *testing.T) {
	defer reset("TestSendMultiCast")
	pb := &testtypes.TestProto{}
	err := eg4.Multicast(types.CastMode_All, types.Action_POST, 0, infra.TEST_TOPIC, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	sleep()
	sleep()

	for eg, tsp := range tsps {
		if tsp.PostN() != 1 && eg != "eg5" {
			Log.Fail(t, eg, " Post count does not equal 1")
			return
		} else if tsp.PostN() != 0 && eg == "eg5" {
			Log.Fail(t, eg, " Post count does not equal 0")
			return
		}
	}
}

func TestUniCast(t *testing.T) {
	defer reset("TestUniCast")
	pb := &testtypes.TestProto{}
	err := eg2.Unicast(types.Action_POST, eg3.Resources().Config().LocalUuid, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostN() != 1 {
		Log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}
}

func TestReconnect(t *testing.T) {
	defer reset("TestReconnect")
	pb := &testtypes.TestProto{}
	err := eg5.Unicast(types.Action_POST, eg3.Resources().Config().LocalUuid, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostN() != 1 {
		Log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}

	Log.Info("********* Starting Reconnect Test")

	//Create a larger than max data
	//sending it will disconnect the socket and attempt a reconnect
	data := make([]byte, eg5.Resources().Config().MaxDataSize+1)
	eg5.SendMessage(data)

	err = eg5.Unicast(types.Action_POST, eg3.Resources().Config().LocalUuid, pb)
	err = eg5.Unicast(types.Action_POST, eg3.Resources().Config().LocalUuid, pb)
	err = eg5.Unicast(types.Action_POST, eg3.Resources().Config().LocalUuid, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	sleep()

	if tsps["eg3"].PostN() != 4 {
		Log.Fail(t, "eg3", " Post count does not equal 4 after reconnect ", tsps["eg3"].PostN())
		return
	}
}

func TestDestinationUnreachable(t *testing.T) {
	defer reset("TestDestinationUnreachable")
	pb := &testtypes.TestProto{}
	err := eg2.Unicast(types.Action_POST, eg4.Resources().Config().LocalUuid, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg4"].PostN() != 1 {
		Log.Fail(t, "eg4", " Post count does not equal 1")
		return
	}

	Log.Info("********* Shutting Down")
	eg4.Shutdown()

	sleep()

	err = eg2.Unicast(types.Action_POST, eg4.Resources().Config().LocalUuid, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	sleep()
	if tsps["eg2"].FailedN() != 1 {
		Log.Fail(t, "eg2", " Fail count does not equal 1")
		return
	}

	h := health.Health(eg2.Resources())
	eg4h := h.HealthPoint(eg4.Resources().Config().LocalUuid)
	if eg4h.Status != types.HealthState_Down {
		Log.Fail(t, "eg4 state", " Not Down")
		return
	}
}
