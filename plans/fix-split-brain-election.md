# Fix: Split-Brain Election Due to Pre-Connection Health Activation

## Bug Summary

`NewVirtualNetworkInterface` calls `health.Activate(vnic, false)` in the constructor (line 119) **before** `Start()` and `WaitForConnection()`. The VNic has `running=false` and no TCP connection at this point.

Health activation triggers `triggerElections`, which sends ServiceRegister, ServiceQuery multicasts and schedules an election via the debouncer. All multicasts silently fail with "Port is not active" because there's no connection.

Each node (VNet, backend, UI) runs its election in isolation, elects itself as leader, and the split brain is permanent.

## Fix: Queue Elections at the VNic Level

When `triggerElections` is called on a VNic that is not yet connected, queue the network operations. Flush the queue at the end of `WaitForConnection()`.

## Changes (in dependency order)

### 1. l8types — Add two methods to IVNic interface

**File**: `go/ifs/VNic.go`

Add to the `IVNic` interface:

```go
// Connected returns true if the VNic has an active network connection.
Connected() bool

// EnqueueElection queues a function to run when the VNic connects.
// If already connected, runs immediately.
EnqueueElection(fn func())
```

---

### 2. l8bus — Implement on VirtualNetworkInterface

**File**: `go/overlay/vnic/VirtualNetworkInterface.go`

Add fields to the struct:

```go
pendingElections []func()
pendingMtx       sync.Mutex
```

Add methods:

```go
func (this *VirtualNetworkInterface) Connected() bool {
    return this.connected
}

func (this *VirtualNetworkInterface) EnqueueElection(fn func()) {
    this.pendingMtx.Lock()
    defer this.pendingMtx.Unlock()
    if this.connected {
        go fn()
        return
    }
    this.pendingElections = append(this.pendingElections, fn)
}

func (this *VirtualNetworkInterface) flushPendingElections() {
    this.pendingMtx.Lock()
    pending := this.pendingElections
    this.pendingElections = nil
    this.pendingMtx.Unlock()
    for _, fn := range pending {
        fn()
    }
}
```

In `WaitForConnection()`, after the connection is established (after the polling loop completes), call:

```go
this.flushPendingElections()
```

---

### 3. l8bus — Implement on VnicVnet

**File**: `go/overlay/vnet/VnicVnet.go`

VnicVnet is always connected (it IS the VNet):

```go
func (this *VnicVnet) Connected() bool {
    return true
}

func (this *VnicVnet) EnqueueElection(fn func()) {
    fn()
}
```

---

### 4. l8services — Modify triggerElections to defer network operations

**File**: `go/services/manager/ServiceActivate.go`

In `triggerElections` (lines 172-197), keep the local participant registration immediate but defer the network operations:

```go
func (this *ServiceManager) triggerElections(serviceName string, serviceArea byte, groupName string, handler ifs.IServiceHandler, vnic ifs.IVNic) {
    _, isMapReduceService := handler.(ifs.IMapReduceService)
    if isMapReduceService {
        fmt.Println("Map Reduce Service:", reflect.ValueOf(handler).Elem().Type().Name())
    }

    groupArea := serviceArea
    if groupName != serviceName {
        groupArea = 0
    }

    // Local participant registration — no network needed, always immediate
    localUuid := this.resources.SysConfig().LocalUuid
    this.participantRegistry.RegisterParticipant(groupName, groupArea, localUuid)

    // Network operations — defer if VNic is not connected yet
    vnic.EnqueueElection(func() {
        vnic.Multicast(serviceName, serviceArea, ifs.ServiceRegister, nil)
        vnic.Multicast(serviceName, serviceArea, ifs.ServiceQuery, nil)
        this.electionDebouncer.RequestElection(groupName, groupArea, vnic)
    })
}
```

---

### 5. Vendor updates

```bash
# In l8bus (picks up l8types)
cd l8bus/go && go mod vendor

# In l8services (picks up l8types + l8bus)
cd l8services/go && go mod vendor

# In downstream projects (l8id, l8erp, etc.)
cd l8id/go && go mod vendor
```

---

## What is NOT changed

- `publishService` and `NotifyServiceAdded` multicasts in `Activate()` also fail pre-connection, but they are re-sent when the VNet discovers the new connection (via handshake, health propagation, and `sendEndPoints`). Only the election is critical because a stale self-election is permanent.
- `RecoveryCheck` fires 5 seconds after activation. By then, `WaitForConnection` has completed and the connection is live. It just needs a valid leader, which this fix provides.

## Verification

After the fix, the log should show:
- **One** leader UUID for Health across all nodes (not three different self-elected leaders)
- No "Port is not active" errors for Health sync
- The "Sync: Health area 0 leader=X localUuid=Y" line should show leader != localUuid on non-leader nodes
