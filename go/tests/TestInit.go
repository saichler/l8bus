package tests

import (
	"crypto/md5"
	"encoding/base64"
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/layer8/go/overlay/switching"
	. "github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/share/logger"
	"github.com/saichler/shared/go/share/service_points"
	"github.com/saichler/shared/go/share/shallow_security"
	"github.com/saichler/shared/go/share/type_registry"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/tests/infra"
	"time"
)

const (
	DEFAULT_MAX_DATA_SIZE     = 1024 * 1024
	DEFAULT_EDGE_QUEUE_SIZE   = 10000
	DEFAULT_SWITCH_QUEUE_SIZE = 50000
	DEFAULT_SWITCH_PORT       = 50000
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
	SetLogger(logger.NewLoggerImpl(&logger.FmtLogMethod{}))
	Logger().SetLogLevel(Trace_Level)
}

func setupTopology() {
	sw1 = createSwitch(50000)
	sw2 = createSwitch(50001)
	eg1 = createEdge(50000, "eg1", true)
	eg2 = createEdge(50000, "eg2", true)
	eg3 = createEdge(50001, "eg3", true)
	eg4 = createEdge(50001, "eg4", true)
	eg5 = createEdge(50000, "eg5", false)
	sleep()
	connectSwitches(sw1, sw2)
	sleep()
	eg1.PublishState()
	eg2.PublishState()
	eg3.PublishState()
	eg4.PublishState()
	sleep()
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

func createSecurityProvider() ISecurityProvider {
	hash := md5.New()
	secret := "Default Security Provider"
	hash.Write([]byte(secret))
	kHash := hash.Sum(nil)
	k := base64.StdEncoding.EncodeToString(kHash)
	return shallow_security.NewShallowSecurityProvider(k, secret)
}

func createProviders() *Providers {
	providers := NewProviders(
		type_registry.NewTypeRegistry(),
		createSecurityProvider(),
		service_points.NewServicePoints(),
		logger.NewLoggerImpl(&logger.FmtLogMethod{}))
	a := NewMessageConfig(DEFAULT_MAX_DATA_SIZE, DEFAULT_EDGE_QUEUE_SIZE,
		DEFAULT_EDGE_QUEUE_SIZE, DEFAULT_SWITCH_PORT, true, 30)
	b := NewMessageConfig(DEFAULT_MAX_DATA_SIZE, DEFAULT_EDGE_QUEUE_SIZE,
		DEFAULT_EDGE_QUEUE_SIZE, DEFAULT_SWITCH_PORT, false, 0)
	c := NewMessageConfig(DEFAULT_MAX_DATA_SIZE, DEFAULT_SWITCH_QUEUE_SIZE,
		DEFAULT_SWITCH_QUEUE_SIZE, DEFAULT_SWITCH_PORT, true, 30)
	providers.SetDefaultMessageConfig(a, c, b)
	return providers
}

func createSwitch(port uint32) *switching.SwitchService {
	providers := createProviders()
	swc := providers.Switch()
	swConfig := &swc
	swConfig.SwitchPort = port
	sw := switching.NewSwitchService(swConfig, providers)
	sw.Start()
	return sw
}

func createEdge(port uint32, name string, addTestTopic bool) IEdge {
	providers := createProviders()
	egc := providers.EdgeConfig()
	egConfig := &egc
	tsps[name] = infra.NewTestServicePointHandler(name)
	providers.ServicePoints().RegisterServicePoint(&tests.TestProto{}, tsps[name], providers.Registry())

	eg, err := edge.ConnectTo("127.0.0.1", port, nil, nil, egConfig, providers)
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

func sleep() {
	time.Sleep(time.Millisecond * 100)
}
