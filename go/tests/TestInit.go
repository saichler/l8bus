package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/switching"
	"github.com/saichler/shared/go/share/defaults"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/service_points"
	"github.com/saichler/shared/go/share/type_registry"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"time"
)

var sw1 *switching.SwitchService
var sw2 *switching.SwitchService
var eg1 IEdge
var eg2 IEdge
var eg3 IEdge
var eg4 IEdge
var eg5 IEdge
var tsps = make(map[string]*infra.TestServicePointHandler)

func init() {
	defaults.LoadDefaultImplementations()
	Logger().SetLogLevel(Trace_Level)
	//Logger().EnableLoggerSync(true)
	setupTopology()
}

func setupTopology() {
	sw1 = createSwitch(50000)
	sw2 = createSwitch(50001)
	eg1 = createEdge(50000, "eg1", true)
	eg2 = createEdge(50000, "eg2", true)
	eg3 = createEdge(50001, "eg3", true)
	eg4 = createEdge(50001, "eg4", true)
	eg5 = createEdge(50000, "eg5", false)
	time.Sleep(time.Second)
	connectSwitches(sw1, sw2)
	time.Sleep(time.Second)
	eg1.PublishState()
	eg2.PublishState()
	eg3.PublishState()
	eg4.PublishState()
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
	swRegistry := type_registry.NewTypeRegistry()
	swServicePoints := service_points.NewServicePoints()
	sw := switching.NewSwitchService(swConfig, swRegistry, swServicePoints)
	sw.Start()
	return sw
}

func createEdge(port uint32, name string, addTestTopic bool) IEdge {
	egConfig := EdgeConfig()
	egRegistry := type_registry.NewTypeRegistry()
	egServicePoints := service_points.NewServicePoints()
	tsps[name] = infra.NewTestServicePointHandler(name)
	egServicePoints.RegisterServicePoint(&tests.TestProto{}, tsps[name], egRegistry)

	eg, err := edge.ConnectTo("127.0.0.1", port, nil, egRegistry, egServicePoints, egConfig)
	if err != nil {
		panic(err.Error())
	}
	if addTestTopic {
		eg.RegisterTopic(infra.TEST_TOPIC)
	}
	eg.(*edge.EdgeImpl).SetName(name)
	return eg
}

func connectSwitches(s1, s2 *switching.SwitchService) {
	s1.ConnectTo("127.0.0.1", s2.Config().SwitchPort)
}
