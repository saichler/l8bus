package tests

import (
	"fmt"
	edge2 "github.com/saichler/overlayK8s/go/overlay/edge"
	"github.com/saichler/overlayK8s/go/overlay/switching"
	"github.com/saichler/shared/go/share/defaults"
	. "github.com/saichler/shared/go/share/interfaces"
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
	sw2 := switching.NewSwitchService(swConfig2, StructRegistry(), ServicePoints())
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
	time.Sleep(time.Millisecond * 100)

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

	fmt.Println(swConfig.Uuid)
	fmt.Println(swConfig2.Uuid)
	fmt.Println(edge1Config.Uuid)

	time.Sleep(time.Second * 10)

}
