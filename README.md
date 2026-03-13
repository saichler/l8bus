# ***"Work less, Achieve much more."***
At first glance, this sounds like a **paradox**—and for most professions,
it probably is. However, in the field of **Software Development**,
I've observed and demonstrated that this apparent contradiction can actually
be resolved into a **Win/Win** situation. By following a few key principles or "rules of thumb," **it's not
only possible but practical**.

[Layer 8 Ecosystem Overview Slide Deck](https://docs.google.com/presentation/d/e/2PACX-1vR7UtPNXRou5uORi-wxZgEYDdVDddT9QCwLH7hrFwnDWJVCx3iCjo6SalAt_jKokB9i_W7mPNU2ntBM/pub?start=false&loop=false&delayms=3000)
# A major lose-lose situation
When microservices started gaining traction around 2013, the promise of modular, scalable
systems captured the industry's attention. However, the lack of mature design patterns,
infrastructure, and tooling—particularly for internal service-to-service integration—introduced
major challenges. Engineers struggled with increased complexity, affecting both delivery
timelines and work/life balance, while organizations faced skyrocketing infrastructure
costs to maintain stability.

Microservices integration typically involves navigating 20 to 30 distinct challenge areas,
each presenting unique failure modes. A misstep in any of these can lead to significant
operational overhead. Engineers are forced into time-consuming maintenance, while businesses
absorb substantial infrastructure costs—resulting in inefficiency and resource drain on both
fronts. **Everyone loses**.

# Introducing Layer8
The architectural and operational principles required for building an "as a service" component
are well-established. However, these principles are often partially overlooked during initial
development—leading to fragile systems and costly outcomes.

**Layer8** is a collection of projects that distills the essential "as a service" requirements
into modular, interoperable building blocks. These components are designed to be agnostic of
one another, enabling teams to rapidly assemble a robust and maintainable microservices
foundation with confidence and high speed.

In other words, by removing the infrastructure burdens of building "as a service" components,
you free yourself to focus on what truly matters—delivering functional value. This shift
allows you to **work less** on boilerplate and operational overhead, and **achieve much more**
by investing your energy in actual product features and business logic.
**A.K.A, Work less, do much more.**

# Precision Matters: Layer8's "As a Service" Offerings
When it comes to microservices, the devil truly is in the details. Layer8 addresses these
intricacies by offering a comprehensive set of "as a service" components that abstract and
standardize foundational concerns:

## Core Infrastructure Services (As a Service):

- Security & AAA (Authentication, Authorization, Accounting)
- Networking
- Messaging
- Health Monitoring
- API Management
- Transactions & Concurrency
- Data Sharding
- Caching
- Serialization
- Model-Agnostic Analysis & Updates
- Query Language Support
- Object-Relational Mapping (ORM)
- Web API Interfaces
- Testing & Quality Assurance
- Collection Management
- Parsing
- Data Modeling

## Layer8 Component Library
Layer8 also provides a modular set of components to enhance service development efficiency
and reliability:

- Logging
- Model Registry
- Synchronized Queues
- Synchronized Maps
- From/To String-Based Utilities

All services and components are **fully decoupled and interface-driven**, ensuring that
applications remain independent of Layer8's specific implementations.
This design allows any service to be easily swapped or replaced if it no longer
meets requirements—without impacting the rest of the system.

# Project Collection
Layer8 is not a single project, but rather a comprehensive ecosystem—a collection of
independent projects, each fulfilling one or more core "As a Service" requirements through
an agnostic, interface-based approach.

## Below is a list of the current Layer8 projects:

- [Layer8 Network & Messaging as a Service](https://github.com/saichler/layer8)
- [Layer8 Service Interfaces](https://github.com/saichler/l8types)
- [Shared Basic Components](https://github.com/saichler/l8utils)
- Security – Private repository (lightweight security implementation available in l8utils for testing)
- [Serialization & Protobuf Tools](https://github.com/saichler/l8srlz)
- [Model Agnostic Analysis & Updating Tools](https://github.com/saichler/reflect)
- [Service API, Concurrency, and Distributed Cache](https://github.com/saichler/l8services)
- [Agnostic Query Language for Models](https://github.com/saichler/gsql)
- [Object Relation Mappings (ORM) as a Service](https://github.com/saichler/l8orm)
- [Web Services for services](https://github.com/saichler/l8web)
- [Testing & Quality Framework](https://github.com/saichler/l8test)
- [Collection, Parsing & Modeling Services](https://github.com/saichler/collect)

Example Application
[Probler – A Kubernetes-native example app built on the Layer8 platform](https://github.com/saichler/probler)

Step by Step Application Development Guide
TBD - work in progress.

## Go Implementation

This repository contains the Go implementation of the Layer8 overlay network in the `/go` directory.

### Core Features
- **Virtual Network Overlay**: TCP-based virtual networking layer with VNet switches and VNic interfaces
- **Smart Service Routing**: Multiple routing algorithms — Proximity, Round Robin, Local, and Leader — for flexible load distribution
- **Multicast Messaging**: Full multicast support for efficient one-to-many communication patterns
- **MapReduce Framework**: Integrated MapR capabilities for distributed data processing and parallel computation
- **Service Discovery**: UDP-based automatic peer discovery and service registration
- **Health Monitoring**: Distributed health scoring with CPU/memory tracking, participant discovery, and leader election
- **Circuit Breaker**: Three-state circuit breaker pattern (Closed → Open → Half-Open) with configurable failure thresholds
- **Metrics Collection**: Real-time performance monitoring with counters, gauges, and histograms
- **Connection Management**: Automatic reconnection, connection pooling, and external NIC support
- **Property Change Notifications**: Event-driven notification system for service state changes
- **Route Table Management**: Dynamic route propagation across multi-VNet topologies
- **Service Groups**: System-level service grouping for coordinated service management

### Architecture

```
l8bus/go/
├── overlay/
│   ├── vnet/        # Virtual Network Switch (12 files)
│   │   ├── VNet.go              - Central switch, TCP listener, lifecycle management
│   │   ├── SwitchTable.go       - Connection + service + route table aggregation
│   │   ├── Connect.go           - VNic connection handling and handshake
│   │   ├── Connections.go       - Connection pool (internal VNics + external VNets)
│   │   ├── Services.go          - Service registry and routing lookups
│   │   ├── RouteTable.go        - Cross-VNet route management
│   │   ├── Discovery.go         - UDP peer discovery
│   │   ├── Notifications.go     - Property change + route broadcast
│   │   ├── VnetHealth.go        - Health record management per VNic
│   │   ├── VnetService.go       - Serialized VNet service request processing
│   │   ├── VNetSystem.go        - System message handling (routes, services)
│   │   └── VnicVnet.go          - VNet-side VNic connection wrapper
│   │
│   ├── vnic/        # Virtual Network Interface (12 files)
│   │   ├── VirtualNetworkInterface.go - Main VNic struct and lifecycle
│   │   ├── API.go               - Public API (Unicast, Multicast, Request)
│   │   ├── SendAlgo.go          - Routing algorithms (Proximity, RoundRobin, Local, Leader)
│   │   ├── SendUnicast.go       - Point-to-point message delivery
│   │   ├── SendMulticast.go     - One-to-many message delivery
│   │   ├── SendForward.go       - Multi-hop message forwarding
│   │   ├── TX.go                - Outbound message serialization and send
│   │   ├── RX.go                - Inbound message receive and dispatch
│   │   ├── KeepAlive.go         - Multicast-based keep-alive heartbeats
│   │   ├── HealthStatistics.go  - Per-VNic health stats reporting
│   │   ├── Notifications.go     - VNic-level notification handling
│   │   └── SubComponents.go     - Internal component initialization
│   │
│   ├── health/      # Health & Leader Election (4 files)
│   │   ├── HealthService.go         - Health service (CRUD + cache)
│   │   ├── HealthServiceCallback.go - Merge logic, leader election, notifications
│   │   ├── HealthStats.go           - CPU/memory tracking via /proc
│   │   └── RoundRobin.go            - Round-robin participant selection
│   │
│   ├── metrics/     # Performance Monitoring (3 files)
│   │   ├── MetricsCollector.go  - Registry with counters, gauges, histograms
│   │   ├── CircuitBreaker.go    - Circuit breaker pattern + manager
│   │   └── ConnectionHealth.go  - Connection-level health scoring
│   │
│   ├── protocol/    # Message Serialization (2 files)
│   │   ├── Protocol.go          - Message creation, serialization, deserialization
│   │   └── MessageCount.go      - Per-service message counting
│   │
│   └── plugins/     # Dynamic Plugin Loading (2 files)
│       ├── PluginCenter.go      - .so plugin discovery and loading
│       └── PluginService.go     - Plugin service activation and lifecycle
│
└── tests/           # Integration Tests (8 test files + init)
```

### Routing Algorithms

The VNic provides multiple message delivery strategies:

| Algorithm | Method | Description |
|-----------|--------|-------------|
| **Proximity** | `Proximity()` / `ProximityRequest()` | Routes to the nearest service instance (same VNet segment) |
| **Round Robin** | `RoundRobin()` / `RoundRobinRequest()` | Distributes requests evenly across all service instances |
| **Local** | `Local()` / `LocalRequest()` | Routes to a service instance on the local VNic only |
| **Leader** | `Leader()` / `LeaderRequest()` | Routes to the leader instance (earliest registered) |
| **Unicast** | `Unicast()` / `Request()` | Direct point-to-point delivery to a specific VNic UUID |
| **Multicast** | `Multicast()` | Broadcasts to all instances of a service |

Each algorithm has a fire-and-forget variant and a synchronous `Request` variant that waits for a response.

### Health & Monitoring

- **CPU Tracking**: Samples `/proc/self/stat` and `/proc/stat` to calculate per-process CPU usage
- **Memory Tracking**: Reports `runtime.MemStats.Alloc` for current heap allocation
- **Health States**: Up / Down status per node, with start time and service list
- **Leader Election**: Deterministic leader selection based on earliest registered participant
- **Circuit Breaker**: Configurable failure threshold (default 5), reset timeout (30s), concurrency limit (100), and success threshold (3) for half-open recovery

### Dependencies
- Go 1.25.4
- Layer8 ecosystem libraries (l8types, l8utils, l8services, l8srlz, l8test)
- Google UUID for unique identifiers
- Protocol Buffers for serialization

### Project Statistics
- **35 source files**, ~4,900 lines of Go code (excluding vendor and tests)
- **8 test files** with integration tests covering overlay, messaging, keep-alive, leader election, scaling, and external VNic scenarios
- Licensed under Apache 2.0

### Building & Testing
```bash
cd go
go build ./...          # Compile all packages
./test.sh               # Run unit/integration tests with coverage
./scale-test.sh         # Run scale/load tests
```

# Detail documenting is WIP...
