package tests

import (
	edge2 "github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/switching"
	"github.com/saichler/shared/go/share/defaults"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/service_points"
	"testing"
	"time"
)

func init() {
	defaults.LoadDefaultImplementations()
}

func TestOverlay(t *testing.T) {
	swConfig := SwitchConfig()
	sw := switching.NewSwitchService(swConfig, StructRegistry(), ServicePoints())
	sw.Start()

	swConfig2 := SwitchConfig()
	swConfig2.SwitchPort = 50001
	sw2ServicePoints := service_points.NewServicePoints()
	sw2 := switching.NewSwitchService(swConfig2, StructRegistry(), sw2ServicePoints)
	sw2.Start()

	defer func() {
		sw.Shutdown()
		sw2.Shutdown()
	}()

	err := sw2.ConnectTo("127.0.0.1", swConfig.SwitchPort)
	if err != nil {
		Fail(t, err)
		return
	}
	time.Sleep(time.Second)

	Info("****************************************************************")

	edge1Config := EdgeConfig()
	eg1, err := edge2.ConnectTo("127.0.0.1", swConfig.SwitchPort, nil, StructRegistry(), ServicePoints(), edge1Config)
	if err != nil {
		Fail(t, err)
		return
	}

	defer func() {
		eg1.Shutdown()
		time.Sleep(time.Second)
	}()

	Info("Switch 1:", swConfig.Local_Uuid)
	Info("Switch 2:", swConfig2.Local_Uuid)
	Info("Edge 1:", edge1Config.Local_Uuid)

	time.Sleep(time.Second * 10)
	Info("**********************************************************")
}
