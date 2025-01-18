package tests

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/switching"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/logger"
	"github.com/saichler/shared/go/share/resources"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"time"
)

var log = logger.NewLoggerImpl(&logger.FmtLogMethod{})
var sw1 *switching.SwitchService
var sw2 *switching.SwitchService
var eg1 IVirtualNetworkInterface
var eg2 IVirtualNetworkInterface
var eg3 IVirtualNetworkInterface
var eg4 IVirtualNetworkInterface
var eg5 IVirtualNetworkInterface
var tsps = make(map[string]*infra.TestServicePointHandler)

func init() {
	log.SetLogLevel(Trace_Level)
	protocol.UsingContainers = false
}

func setupTopology() {
	sw1 = createSwitch(50000, "sw1")
	sw2 = createSwitch(50001, "sw2")
	sleep()
	eg1 = createEdge(50000, "eg1", true)
	eg2 = createEdge(50000, "eg2", true)
	eg3 = createEdge(50001, "eg3", true)
	eg4 = createEdge(50001, "eg4", true)
	eg5 = createEdge(50000, "eg5", false)
	sleep()
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
	sleep()
}

func createSwitch(port uint32, name string) *switching.SwitchService {
	res := resources.NewDefaultResources(log, name)
	res.Config().SwitchPort = port
	sw := switching.NewSwitchService(res)
	sw.Start()
	return sw
}

func createEdge(port uint32, name string, addTestTopic bool) IVirtualNetworkInterface {
	resources := resources.NewDefaultResources(log, name)
	resources.Config().SwitchPort = port
	tsps[name] = infra.NewTestServicePointHandler(name)

	if addTestTopic {
		sp := resources.ServicePoints()
		err := sp.RegisterServicePoint(&tests.TestProto{}, tsps[name])
		if err != nil {
			panic(err)
		}
	}

	vnic := vnic2.NewVirtualNetworkInterface(resources, nil)
	vnic.Start()

	/*
		if addTestTopic {

			eg.RegisterTopic(infra.TEST_TOPIC)
		}
		eg.(*edge.EdgeImpl).SetName(name)

	*/
	return vnic
}

func connectSwitches(s1, s2 *switching.SwitchService) {
	s1.Switch2Switch("127.0.0.1", s2.Resources().Config().SwitchPort)
}

func sleep() {
	time.Sleep(time.Millisecond * 100)
}
