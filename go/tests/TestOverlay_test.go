package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_servicepoints"
	. "github.com/saichler/l8test/go/infra/t_topology"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func isVnic1Ready() bool {
	nic := topo.VnicByVnetNum(1, 1)
	hc := health.Health(nic.Resources())
	hp := hc.All()
	if len(hp) != 15 {
		return false
	}
	return true
}

func TestTopologyHealth(t *testing.T) {
	if !WaitForCondition(isVnic1Ready, 5, t, "Vnic1 is not ready") {
		return
	}
	defer reset("TestTopologyHealth")
	for vnetNum := 1; vnetNum <= 3; vnetNum++ {
		for vnicNum := 1; vnicNum <= 4; vnicNum++ {
			nic := topo.VnicByVnetNum(vnetNum, vnicNum)
			hc := health.Health(nic.Resources())
			hp := hc.All()
			if len(hp) != 15 {
				Log.Fail(t, "Expected ", nic.Resources().SysConfig().LocalAlias,
					" to have 15 heath points, but it has ", len(hp))
				for _, h := range hp {
					Log.Info(h.Alias)
				}
				return
			}
		}
	}
	eg1_1 := topo.VnicByVnetNum(1, 1)
	eg2_1 := topo.VnicByVnetNum(2, 1)
	eg3_1 := topo.VnicByVnetNum(3, 1)
	eg1_1_Points := health.Health(eg1_1.Resources()).All()
	eg2_1_Points := health.Health(eg2_1.Resources()).All()
	eg3_1_Points := health.Health(eg3_1.Resources()).All()
	if len(eg1_1_Points) != len(eg2_1_Points) || len(eg1_1_Points) != len(eg3_1_Points) {
		Log.Fail(t, "Expected health points to be equal ", len(eg1_1_Points), ":", len(eg2_1_Points), ":", len(eg3_1_Points))
		return
	}
	for k, _ := range eg1_1_Points {
		delete(eg2_1_Points, k)
	}
	if len(eg2_1_Points) != 0 {
		Log.Fail(t, "Expected health points to be empty ", len(eg1_1_Points), ":", len(eg2_1_Points), ":", len(eg3_1_Points))
		return
	}

	hc := health.Health(eg3_1.Resources())
	uuids := hc.Uuids(ServiceName, 0)
	if len(uuids) != 9 {
		Log.Fail(t, "Expected uuids to be 9, but it is ", len(uuids))
		return
	}

	uuids = hc.Uuids(health.ServiceName, 0)
	if len(uuids) != 15 {
		Log.Fail(t, "Expected uuids to be 15, but it is ", len(uuids))
		for uuid, _ := range uuids {
			p := hc.HealthPoint(uuid)
			Log.Info(p.Alias)
		}
		return
	}
}

func TestSendMultiCast(t *testing.T) {
	defer reset("TestSendMultiCast")
	Log.Info("*** Sending Multicast Message")
	pb := CreateTestModelInstance(3)
	eg2_1 := topo.VnicByVnetNum(2, 1)
	err := eg2_1.Multicast(ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	Sleep()
	Sleep()
	handlers := topo.AllHandlers()

	if len(handlers) != 9 {
		Log.Fail(t, "Expected handlers to be 9, but it is ", len(handlers))
		return
	}

	posts := 0
	for _, handler := range handlers {
		posts += handler.PostN()
		if handler.PostN() == 0 {
			Log.Error(handler.Name())
		}
	}

	if posts != 9 {
		Log.Fail(t, "Expected 9 posts but got ", posts)
		return
	}
}

func TestUniCast(t *testing.T) {
	defer reset("TestUniCast")
	pb := CreateTestModelInstance(3)
	eg1_2 := topo.VnicByVnetNum(1, 2)
	eg3_3 := topo.VnicByVnetNum(3, 3)
	err := eg1_2.Unicast(eg3_3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	Sleep()
	handler := topo.HandlerByVnetNum(3, 3)
	if handler.PostN() != 1 {
		Log.Fail(t, "eg3_3", " Post count does not equal 1")
		return
	}
}

func TestReconnect(t *testing.T) {
	defer reset("TestReconnect")
	pb := CreateTestModelInstance(3)
	eg2_1 := topo.VnicByVnetNum(2, 1)
	eg1_3 := topo.VnicByVnetNum(1, 3)
	err := eg2_1.Unicast(eg1_3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	Sleep()
	handler := topo.HandlerByVnetNum(1, 3)
	if handler.PostN() != 1 {
		Log.Fail(t, "eg3_1", " Post count does not equal 1")
		return
	}

	Log.Info("********* Starting Reconnect Test")
	//Create a larger than max data
	//sending it will disconnect the socket and attempt a reconnect
	data := make([]byte, eg2_1.Resources().SysConfig().MaxDataSize+1)
	eg2_1.SendMessage(data)

	err = eg2_1.Unicast(eg1_3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	err = eg2_1.Unicast(eg1_3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	err = eg2_1.Unicast(eg1_3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	Sleep()

	if handler.PostN() != 4 {
		Log.Fail(t, "eg3", " Post count does not equal 4 after reconnect ", handler.PostN())
		return
	}
}

func TestDestinationUnreachable(t *testing.T) {
	defer reset("TestDestinationUnreachable")
	pb := CreateTestModelInstance(3)
	eg3_2 := topo.VnicByVnetNum(3, 2)
	eg1_1 := topo.VnicByVnetNum(1, 1)
	defer func() {
		topo.RenewVnic(eg1_1.Resources().SysConfig().LocalAlias)
	}()

	err := eg3_2.Unicast(eg1_1.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	Sleep()

	handler := topo.HandlerByVnetNum(1, 1)

	if handler.PostN() != 1 {
		Log.Fail(t, "eg1_1", " Post count does not equal 1 ", handler.PostN())
		return
	}

	Log.Info("********* Shutting Down")
	eg1_1.Shutdown()

	Sleep()

	err = eg3_2.Unicast(eg1_1.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	Sleep()
	handler = topo.HandlerByVnetNum(3, 2)

	if handler.FailedN() != 1 {
		Log.Fail(t, "eg3_2", " Fail count does not equal 1")
		return
	}

	h := health.Health(eg3_2.Resources())
	eg4h := h.HealthPoint(eg1_1.Resources().SysConfig().LocalUuid)
	if eg4h.Status != types.HealthState_Down {
		Log.Fail(t, "eg1_1 state", " Not Down")
		return
	}
}
