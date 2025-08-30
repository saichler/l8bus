package vnic

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/l8utils/go/utils/strings"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/metrics"
	"github.com/saichler/layer8/go/overlay/plugins"
	"github.com/saichler/layer8/go/overlay/protocol"
	requests2 "github.com/saichler/layer8/go/overlay/vnic/requests"
)

type VirtualNetworkInterface struct {
	// Resources for this VNic such as registry, security & config
	resources ifs.IResources
	// The socket connection
	conn net.Conn
	// The socket connection mutex
	connMtx *sync.Mutex
	// is running
	running bool
	// Sub components/go routines
	components *SubComponents
	// The Protocol
	protocol *protocol.Protocol
	// Name for this VNic expressing the connection path in aside -->> zside
	name string
	// Indicates if this vnic in on the switch internal, hence need no keep alive
	IsVNet bool
	// Last reconnect attempt
	last_reconnect_attempt int64

	requests *requests2.Requests

	healthStatistics  *HealthStatistics
	connectionMetrics *metrics.ConnectionMetrics
	circuitBreaker    *metrics.CircuitBreaker
	metricsRegistry   *metrics.MetricsRegistry
	connected         bool
}

func NewVirtualNetworkInterface(resources ifs.IResources, conn net.Conn) *VirtualNetworkInterface {
	vnic := &VirtualNetworkInterface{}
	vnic.conn = conn
	vnic.resources = resources
	vnic.connMtx = &sync.Mutex{}
	vnic.protocol = protocol.New(resources)
	vnic.components = newSubomponents()
	vnic.components.addComponent(newRX(vnic))
	vnic.components.addComponent(newTX(vnic))
	vnic.components.addComponent(newKeepAlive(vnic))
	vnic.requests = requests2.NewRequests()
	vnic.healthStatistics = &HealthStatistics{}
	
	// Initialize metrics system
	vnic.metricsRegistry = metrics.GetGlobalRegistry(resources.Logger())
	
	// Initialize connection metrics if we have a connection
	if conn != nil {
		remoteAddr := conn.RemoteAddr().String()
		connectionID := ifs.NewUuid()
		vnic.connectionMetrics = metrics.NewConnectionMetrics(connectionID, remoteAddr)
		
		// Initialize circuit breaker for this connection
		cbManager := metrics.NewCircuitBreakerManager(vnic.metricsRegistry, resources.Logger())
		cbConfig := metrics.DefaultCircuitBreakerConfig()
		vnic.circuitBreaker = cbManager.GetOrCreate("vnic_"+connectionID, cbConfig)
	}
	
	if vnic.resources.SysConfig().LocalUuid == "" {
		vnic.resources.SysConfig().LocalUuid = ifs.NewUuid()
	}

	if conn == nil {
		// Register the health service
		vnic.resources.Services().RegisterServiceHandlerType(&health.HealthService{})
		vnic.resources.Services().Activate(health.ServiceTypeName, health.ServiceName, 0, vnic.resources, nil)
		vnic.resources.Services().RegisterServiceHandlerType(&plugins.PluginService{})
		vnic.resources.Services().Activate(plugins.ServiceTypeName, plugins.ServiceName, 0, vnic.resources, nil)
	}

	return vnic
}

func (this *VirtualNetworkInterface) Start() {
	this.running = true
	if this.conn == nil {
		this.connectToSwitch()
	} else {
		this.receiveConnection()
	}
	this.name = strings.New(this.resources.SysConfig().LocalAlias, " -->> ", this.resources.SysConfig().RemoteAlias).String()
}

func (this *VirtualNetworkInterface) connectToSwitch() {
	err := this.connect()
	if err != nil {
		panic(err)
	}
	this.components.start()
	this.connected = true
}

func (this *VirtualNetworkInterface) connect() error {
	// Dial the destination and validate the secret and key
	destination := protocol.MachineIP
	if ifs.NetworkMode_K8s() {
		destination = os.Getenv("NODE_IP")
	} else if ifs.NetworkMode_DOCKER() {
		// inside a containet the switch ip will be the external subnet + ".1"
		// for example if the address of the container is 172.1.1.112, the switch will be accessible
		// via 172.1.1.1
		subnet := protocol.IpSegment.ExternalSubnet()
		destination = strings.New(subnet, ".1").String()
	}
	this.resources.Logger().Info("Trying to connect to vnet at IP - ", destination)
	// Try to dial to the switch
	conn, err := this.resources.Security().CanDial(destination, this.resources.SysConfig().VnetPort)
	if err != nil {
		return errors.New("Error connecting to the vnet: " + err.Error())
	}
	// Verify that the switch accepts this connection
	if this.resources.SysConfig().LocalUuid == "" {
		panic("Couldn't connect")
	}
	this.syncServicesWithConfig()
	err = this.resources.Security().ValidateConnection(conn, this.resources.SysConfig())
	if err != nil {
		return errors.New("Error validating connection: " + err.Error())
	}
	this.conn = conn
	this.resources.SysConfig().Address = conn.LocalAddr().String()
	this.resources.Logger().Info("Connected!")
	return nil
}

func (this *VirtualNetworkInterface) syncServicesWithConfig() {
	s1 := this.resources.Services().Services()
	s2 := this.resources.SysConfig().Services
	if s2 == nil {
		this.resources.SysConfig().Services = s1
		return
	}
	for k, v := range s1.ServiceToAreas {
		for k1, _ := range v.Areas {
			_, ok := s2.ServiceToAreas[k]
			if !ok {
				s2.ServiceToAreas[k] = &types.ServiceAreas{}
				s2.ServiceToAreas[k].Areas = make(map[int32]bool)
			}
			s2.ServiceToAreas[k].Areas[k1] = true
		}
	}
}

func (this *VirtualNetworkInterface) receiveConnection() {
	this.IsVNet = true
	this.resources.SysConfig().Address = this.conn.RemoteAddr().String()
	this.components.start()
}

func (this *VirtualNetworkInterface) Shutdown() {
	this.resources.Logger().Info("Shutdown was called on ", this.resources.SysConfig().LocalAlias)
	this.running = false
	if this.conn != nil {
		this.conn.Close()
	}
	this.components.shutdown()
	if this.resources.DataListener() != nil {
		go this.resources.DataListener().ShutdownVNic(this)
	}
}

func (this *VirtualNetworkInterface) Name() string {
	if this.name == "" {
		this.name = strings.New(this.resources.SysConfig().LocalUuid,
			" -->> ",
			this.resources.SysConfig().RemoteUuid).String()
	}
	return this.name
}

func (this *VirtualNetworkInterface) SendMessage(data []byte) error {
	return this.components.TX().SendMessage(data)
}

func (this *VirtualNetworkInterface) ServiceAPI(serviceName string, serviceArea byte) ifs.ServiceAPI {
	return newAPI(serviceName, serviceArea, false, false)
}

func (this *VirtualNetworkInterface) Resources() ifs.IResources {
	return this.resources
}

func (this *VirtualNetworkInterface) reconnect() {
	this.connMtx.Lock()
	defer this.connMtx.Unlock()
	if !this.running {
		return
	}
	if time.Now().Unix()-this.last_reconnect_attempt < 5 {
		return
	}
	this.last_reconnect_attempt = time.Now().Unix()

	this.resources.Logger().Info("***** Trying to reconnect to ", this.resources.SysConfig().RemoteAlias, " *****")

	if this.conn != nil {
		this.conn.Close()
	}

	err := this.connect()
	if err != nil {
		this.resources.Logger().Error("***** Failed to reconnect to ", this.resources.SysConfig().RemoteAlias, " *****")
	} else {
		this.resources.Logger().Info("***** Reconnected to ", this.resources.SysConfig().RemoteAlias, " *****")
	}
}

func (this *VirtualNetworkInterface) WaitForConnection() {
	for !this.connected {
		time.Sleep(time.Millisecond * 100)
	}
	hc := health.Health(this.resources)
	hp := hc.Health(this.resources.SysConfig().LocalUuid)
	for hp == nil {
		time.Sleep(time.Millisecond * 100)
		hp = hc.Health(this.resources.SysConfig().LocalUuid)
	}
}

func (this *VirtualNetworkInterface) Running() bool {
	return this.running
}

// RecordMessageSent records metrics for an outgoing message
func (this *VirtualNetworkInterface) RecordMessageSent(bytes int64) {
	if this.connectionMetrics != nil {
		this.connectionMetrics.RecordMessageSent(bytes)
	}
	
	// Update global metrics
	if this.metricsRegistry != nil {
		sentCounter := this.metricsRegistry.Counter("layer8_messages_sent_total", 
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		sentCounter.Inc()
		
		bytesCounter := this.metricsRegistry.Counter("layer8_bytes_sent_total",
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		bytesCounter.Add(bytes)
	}
}

// RecordMessageReceived records metrics for an incoming message
func (this *VirtualNetworkInterface) RecordMessageReceived(bytes int64) {
	if this.connectionMetrics != nil {
		this.connectionMetrics.RecordMessageReceived(bytes)
	}
	
	// Update global metrics
	if this.metricsRegistry != nil {
		receivedCounter := this.metricsRegistry.Counter("layer8_messages_received_total",
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		receivedCounter.Inc()
		
		bytesCounter := this.metricsRegistry.Counter("layer8_bytes_received_total",
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		bytesCounter.Add(bytes)
	}
}

// RecordError records a connection error
func (this *VirtualNetworkInterface) RecordError() {
	if this.connectionMetrics != nil {
		this.connectionMetrics.RecordError()
	}
	
	if this.metricsRegistry != nil {
		errorCounter := this.metricsRegistry.Counter("layer8_connection_errors_total",
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		errorCounter.Inc()
	}
}

// RecordLatency records a latency measurement
func (this *VirtualNetworkInterface) RecordLatency(latencyMs int64) {
	if this.connectionMetrics != nil {
		this.connectionMetrics.RecordLatency(latencyMs)
	}
	
	if this.metricsRegistry != nil {
		latencyHistogram := this.metricsRegistry.Histogram("layer8_message_latency_ms",
			map[string]string{"vnic_id": this.resources.SysConfig().LocalUuid})
		latencyHistogram.Observe(latencyMs)
	}
}

// GetConnectionHealth returns the current connection health score
func (this *VirtualNetworkInterface) GetConnectionHealth() int64 {
	if this.connectionMetrics != nil {
		return this.connectionMetrics.GetHealthScore()
	}
	return 100 // Default to healthy if no metrics
}

// GetCircuitBreaker returns the circuit breaker for this connection
func (this *VirtualNetworkInterface) GetCircuitBreaker() *metrics.CircuitBreaker {
	return this.circuitBreaker
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (this *VirtualNetworkInterface) ExecuteWithCircuitBreaker(fn func() (interface{}, error)) (interface{}, error) {
	if this.circuitBreaker != nil {
		return this.circuitBreaker.Execute(fn)
	}
	// Fallback to direct execution if no circuit breaker
	return fn()
}
