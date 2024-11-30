package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
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
	interfaces.Info("Sending data")
	err := eg1.Do(types.Action_POST, infra.TEST_TOPIC, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	time.Sleep(time.Second)

	for eg, tsp := range tsps {
		if tsp.PostNumber != 1 && eg != "eg5" {
			interfaces.Fail(t, eg, " Post count does not equal 1")
			return
		} else if tsp.PostNumber != 0 && eg == "eg5" {
			interfaces.Fail(t, eg, " Post count does not equal 0")
			return
		}
	}

	err = eg2.Do(types.Action_POST, eg3.Config().Local_Uuid, pb)
	if err != nil {
		interfaces.Fail(t, err)
		return
	}
	time.Sleep(time.Second)

	if tsps["eg3"].PostNumber != 2 {
		interfaces.Fail(t, "eg3", " Post count does not equal 2")
	}

	interfaces.Info("*****************************************************************")
	interfaces.Logger().SetLogLevel(interfaces.Info_Level)
	exp := 1002
	scaleTest(1000, exp, 2, t)
	exp += 10000
	scaleTest(10000, exp, 2, t)
	exp += 100000
	scaleTest(100000, exp, 5, t)
	exp += 1000000
	scaleTest(1000000, exp, 5, t)
}

func scaleTest(size, exp int, timeout int64, t *testing.T) {
	start := time.Now().Unix()
	for i := 0; i < size; i++ {
		pb := &tests.TestProto{}
		pb.Int32 = int32(i)
		err := eg2.Do(types.Action_POST, eg3.Config().Local_Uuid, pb)
		if err != nil {
			interfaces.Fail(t, err)
			return
		}
	}

	eg3 := tsps["eg3"]

	now := time.Now().Unix()
	for eg3.PostNumber < exp {
		if time.Now().Unix()-timeout >= now {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	end := time.Now().Unix()
	interfaces.Info("Scale test for ", size, " took ", (end - start), " seconds")
	if eg3.PostNumber != exp {
		interfaces.Fail(t, "eg3", " Post count does not equal to ", exp, ":", eg3.PostNumber)
	}
}
