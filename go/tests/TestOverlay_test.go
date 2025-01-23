package tests

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/shared/go/types"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func reset(name string) {
	log.Info("*** ", name, " end ***")
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
	health.Health(eg5.Resources()).Print()
	health.Health(eg4.Resources()).Print()
	health.Health(sw1.Resources()).Print()
	eg4Points := health.Health(eg4.Resources()).AllPoints()
	eg5Points := health.Health(eg5.Resources()).AllPoints()
	if len(eg5Points) != len(eg4Points) {
		log.Fail("Expected health points to be equal")
		return
	}
	for k, _ := range eg4Points {
		delete(eg5Points, k)
	}
	if len(eg5Points) != 0 {
		log.Fail("Expected health points to be empty")
	}
}

func TestSendMultiCast(t *testing.T) {
	defer reset("TestSendMultiCast")
	pb := &tests.TestProto{}
	err := eg1.Do(types.Action_POST, infra.TEST_TOPIC, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}
	sleep()

	for eg, tsp := range tsps {
		if tsp.PostNumber != 1 && eg != "eg5" {
			log.Fail(t, eg, " Post count does not equal 1")
			return
		} else if tsp.PostNumber != 0 && eg == "eg5" {
			log.Fail(t, eg, " Post count does not equal 0")
			return
		}
	}
}

func TestUniCast(t *testing.T) {
	defer reset("TestUniCast")
	pb := &tests.TestProto{}
	err := eg2.Do(types.Action_POST, eg3.Resources().Config().Local_Uuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostNumber != 1 {
		log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}
}

func TestReconnect(t *testing.T) {
	defer reset("TestReconnect")
	pb := &tests.TestProto{}
	err := eg5.Do(types.Action_POST, eg3.Resources().Config().Local_Uuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg3"].PostNumber != 1 {
		log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}

	log.Info("********* Starting Reconnect Test")

	//Create a larger than max data
	//sending it will disconnect the socket and attempt a reconnect
	data := make([]byte, eg5.Resources().Config().MaxDataSize+1)
	eg5.Send(data)

	err = eg5.Do(types.Action_POST, eg3.Resources().Config().Local_Uuid, pb)
	err = eg5.Do(types.Action_POST, eg3.Resources().Config().Local_Uuid, pb)
	err = eg5.Do(types.Action_POST, eg3.Resources().Config().Local_Uuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}

	sleep()

	if tsps["eg3"].PostNumber != 4 {
		log.Fail(t, "eg3", " Post count does not equal 4 after reconnect")
		return
	}
}

func TestDestinationUnreachable(t *testing.T) {
	defer reset("TestDestinationUnreachable")
	pb := &tests.TestProto{}
	err := eg2.Do(types.Action_POST, eg4.Resources().Config().Local_Uuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}
	sleep()

	if tsps["eg4"].PostNumber != 1 {
		log.Fail(t, "eg4", " Post count does not equal 1")
		return
	}

	log.Info("********* Shutting Down")
	eg4.Shutdown()

	time.Sleep(time.Second * 7)

	err = eg2.Do(types.Action_POST, eg4.Resources().Config().Local_Uuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}
	sleep()
}
