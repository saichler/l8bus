# Layer8 Overlay Network Package

The overlay package provides a comprehensive virtual networking layer built on top of TCP connections, enabling distributed systems to communicate as if they were on a single network.

## Architecture Overview

The overlay network consists of several key components:

### Core Components

#### VNet (Virtual Network)
- **Location**: `vnet/VNet.go`
- **Purpose**: Central network switch that manages connections between nodes
- **Key Features**:
  - TCP listener for incoming connections
  - Message routing and switching
  - Health monitoring integration
  - Service discovery
  - Connection management

#### VNic (Virtual Network Interface)
- **Location**: `vnic/VirtualNetworkInterface.go`
- **Purpose**: Network interface for individual nodes to connect to the overlay
- **Key Features**:
  - Bidirectional communication (TX/RX)
  - Keep-alive mechanisms
  - Automatic reconnection
  - Service API integration

#### Protocol
- **Location**: `protocol/Protocol.go`
- **Purpose**: Message serialization/deserialization and protocol handling
- **Key Features**:
  - Base64 encoded message data
  - Sequence number management
  - Message creation and parsing
  - Transaction state management

## Sub-packages

### Health (`health/`)
- **HealthCenter**: Distributed health monitoring system
- **HealthService**: Service for health status management
- **Services**: Service registry and management
- Tracks node status, service availability, and leader election

### Plugins (`plugins/`)
- **PluginCenter**: Plugin management system
- **PluginService**: Service for loading and managing plugins
- Supports dynamic loading of `.so` files for extending functionality

### Protocol (`protocol/`)
- **Protocol**: Core message handling
- **IPSegment**: IP address management and subnet detection
- **MessageCount**: Message statistics and counting

### VNet (`vnet/`)
- **VNet**: Main virtual network switch
- **Connect**: Connection management utilities
- **Connections**: Internal/external connection tracking
- **Discovery**: Network discovery via UDP broadcasts
- **Notifications**: Event notification system
- **SwitchTable**: Routing table for message forwarding

### VNic (`vnic/`)
- **VirtualNetworkInterface**: Network interface implementation
- **API**: Service API framework
- **KeepAlive**: Connection health monitoring
- **RX/TX**: Receive and transmit components
- **SendMethods**: Message sending utilities
- **SubComponents**: Component lifecycle management
- **requests/**: Request handling framework

## Key Features

### Network Discovery
- Automatic peer discovery using UDP broadcasts
- Dynamic network topology building
- IP-based routing decisions

### Message Routing
- Service-based message routing
- Support for unicast and multicast
- Leader election for service instances
- Transaction state management

### Connection Management
- Internal vs external connection classification
- Automatic reconnection on failures
- Connection pooling and reuse

### Service Framework
- Plugin-based architecture
- Service registration and discovery
- Health monitoring integration
- API abstraction layer

### Security
- Connection validation
- Security provider integration
- Token-based authentication

## Usage

### Creating a VNet (Network Switch)
```go
vnet := NewVNet(resources)
err := vnet.Start()
```

### Creating a VNic (Network Interface)
```go
vnic := NewVirtualNetworkInterface(resources, nil)
vnic.Start()
```

### Service Registration
Services are automatically registered through the health system and can be discovered by other nodes in the network.

## Network Topology

The overlay creates a hybrid network topology:
- **Internal connections**: Direct connections between processes on the same machine
- **External connections**: TCP connections to remote VNet switches
- **Service routing**: Messages routed based on service names and areas

## Dependencies

- `github.com/saichler/l8types`: Type definitions and interfaces
- `github.com/saichler/l8utils`: Utility functions and resources
- `github.com/saichler/l8services`: Service framework
- `github.com/saichler/l8srlz`: Serialization library
- `google.golang.org/protobuf`: Protocol buffer support

## Configuration

The overlay network is configured through `SysConfig` which includes:
- Local and remote UUIDs and aliases
- VNet port configuration
- Service definitions
- Queue sizes and data limits