package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/switching"
	"github.com/saichler/shared/go/share/defaults"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/service_points"
	"github.com/saichler/shared/go/share/struct_registry"
	"time"
)

var sw1 *switching.SwitchService
var sw2 *switching.SwitchService
var eg1 IEdge
var eg2 IEdge
var eg3 IEdge
var eg4 IEdge

func init() {
	defaults.LoadDefaultImplementations()
	setupTopology()
}

func setupTopology() {
	sw1 = createSwitch(50000)
	sw2 = createSwitch(50001)
	eg1 = createEdge(50000)
	eg2 = createEdge(50000)
	eg3 = createEdge(50001)
	eg4 = createEdge(50001)
	time.Sleep(time.Second)
	connectSwitches(sw1, sw2)
	time.Sleep(time.Second)
}

func shutdownTopology() {
	eg4.Shutdown()
	eg3.Shutdown()
	eg2.Shutdown()
	eg1.Shutdown()
	sw2.Shutdown()
	sw1.Shutdown()
	time.Sleep(time.Second)
}

func createSwitch(port uint32) *switching.SwitchService {
	swConfig := SwitchConfig()
	swConfig.SwitchPort = port
	swRegistry := struct_registry.NewStructRegistry()
	swServicePoints := service_points.NewServicePoints()
	sw := switching.NewSwitchService(swConfig, swRegistry, swServicePoints)
	sw.Start()
	return sw
}

func createEdge(port uint32) IEdge {
	egConfig := EdgeConfig()
	egRegistry := struct_registry.NewStructRegistry()
	egServicePoints := service_points.NewServicePoints()
	eg, err := edge.ConnectTo("127.0.0.1", port, nil, egRegistry, egServicePoints, egConfig)
	if err != nil {
		Fail(nil, err)
		return nil
	}
	eg.Start()
	return eg
}

func connectSwitches(s1, s2 *switching.SwitchService) {
	s1.ConnectTo("127.0.0.1", s2.Config().SwitchPort)
}
