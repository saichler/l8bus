package tests

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/layer8/go/overlay/vnet"
	vnic2 "github.com/saichler/layer8/go/overlay/vnic"
	"github.com/saichler/reflect/go/reflect/inspect"
	"github.com/saichler/servicepoints/go/points/service_points"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/logger"
	"github.com/saichler/shared/go/share/registry"
	"github.com/saichler/shared/go/share/resources"
	"github.com/saichler/shared/go/share/shallow_security"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/shared/go/types"
	"time"
)

var log = logger.NewLoggerDirectImpl(&logger.FmtLogMethod{})
var sw1 *vnet.VNet
var sw2 *vnet.VNet
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

func createSwitch(port uint32, name string) *vnet.VNet {
	reg := registry.NewRegistry()
	security := shallow_security.CreateShallowSecurityProvider()
	config := &types.VNicConfig{MaxDataSize: resources.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize: resources.DEFAULT_QUEUE_SIZE,
		TxQueueSize: resources.DEFAULT_QUEUE_SIZE,
		LocalAlias:  name,
		Topics:      map[string]bool{}}
	ins := inspect.NewIntrospect(reg)
	sps := service_points.NewServicePoints(ins, config)

	res := resources.NewResources(reg, security, sps, log, nil, nil, config, ins)
	res.Config().SwitchPort = port
	sw := vnet.NewVNet(res)
	sw.Start()
	return sw
}

func createEdge(port uint32, name string, addTestTopic bool) IVirtualNetworkInterface {
	reg := registry.NewRegistry()
	security := shallow_security.CreateShallowSecurityProvider()
	config := &types.VNicConfig{MaxDataSize: resources.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize: resources.DEFAULT_QUEUE_SIZE,
		TxQueueSize: resources.DEFAULT_QUEUE_SIZE,
		LocalAlias:  name,
		Topics:      map[string]bool{}}
	ins := inspect.NewIntrospect(reg)
	sps := service_points.NewServicePoints(ins, config)

	resources := resources.NewResources(reg, security, sps, log, nil, nil, config, ins)
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

func connectSwitches(s1, s2 *vnet.VNet) {
	s1.ConnectNetworks("127.0.0.1", s2.Resources().Config().SwitchPort)
}

func sleep() {
	time.Sleep(time.Millisecond * 100)
}
