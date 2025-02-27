# Layer 8 - The Missing Development Stack for Micro Services

## Overview

**Micro Services** has exposed a missing developer stack with the Service 2 Service
integration, challenges around **Process 2 Process communication**, **Process 2 Process Security & AAA**,
**Process 2 Process API Sharing**, **Multi Process Stateful Concurrency**, **Service Health Monitorying** & more.

With the above missing standardization & SDK, the **"wheel"** is reinvented every time
and the cost of developing a Micro Service based application are **skyrocketing
and overwhelming the cost of actual business logic code by factors.**

## Layer 8

**Layer8** is a layered development stack & SDK that targets the painful
challenges of Micro Services base **Application** development and trying to encapsulate them into
a very, extremely simple, interface to substantially lower the **Micro Services** development costs
and extremely reduce the **Time to Market**.

**And doing that without settling on Security & Quality**

# Context

1. [Security Provider](#security)
2. [Vnet (Virtual Network)](#vnet)
3. [Vnic (Virtual Network Interface)](#Vnic)
4. [Unicast](#unicast)
5. [Multicast](#multicast)
6. [Unicast Topic](#unicasttopic)
7. [Request/Reply](#requestreply)
8. [Service Points](#servicepoints)
9. [Integrated Health Service & Leader/Follower Election](#health) 
10. [Invoking an API](#api)
11. [Service Transactions](#transactions)
12. [GSQL (Graph SQL)](#gsql)
13. [From String](#fromstring)
14. [Introspector](#introspectpr)
15. [Protobug Object](#object)
16. [Meta Data Driven Property & Dynamic Instantiation](#property)
17. [Updater & Generic Model Change Set](#updater)
18. [Deep Clone (Model Sensitive)](#clone)
19. [Distributed Cache & Delta Notifications](#cache)
20. [Distributed Collection Service](#collect)
21. [Distributed Model Agnostic Parsing](#parse)
22. [Distributed Model Agnostic Inventory](#inventory)
23. [Traffic Generator](#generator)

## Security Provider <a name="security"></a>

When developing a Micro Services stack, usually security consideration comes a bit late
in the game, creating tough challenges. With **Layer8**, this is the starting point!

The **Security Provider** is an abstraction inteface for AAA & Encryption, being used and utilized
the development stack & frameworks. **The Prime rule** is that two component need to have
the **same Security Provider** to interact with each other.

What is the **Same Security Provider?** As this is an interfaced abstraction,
the implementation is currently private...;o)

## Vnet (Virtual Network) <a name="vnet"></a>

**Vnet** is a process running on each host as an OS service,
that we want as part of the **Application Overlay**.
There is the flexability to have multiple Application **Vnet** hosted inside the same OS service
or an OS service per Application Vnet
![alt text](https://github.com/saichler/layer8/blob/main/docs/vnet.png)

## Vnic (Virtual Network Interface) <a name="vnic"></a>

The **Vnic** is a piece of code/library used inside the running process to connect and send/publish/request messages
inside the **Vnet**. When instantiated, it autodetect and connects to the Vnet, **given**
it has the correct **Security Provider**. It is **agnostic** to being hosted inside K8s, Docker, Container or plain
process.
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnic2vnet-connect.png)

## Unicast <a name="unicast"></a>

The **Vnic** can unicast a message to another **Vnic** on the **Vnet** via its uuid address.
Each **Vnic**, once joins the **VNet**, has access to the Health system, via which it can acquire
the uuid of the unicast destination.
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnet-unicast-cross-nodes.png)

## Multicast <a name="multicast"></a>

A Vnic can publish a message to a **Topic**. Any **Vnic** that registered on the **Topic**,
will have the message deliver to it. The **Vnet** on the same Host as the **Vnic**, will forward
the message to its adjacents **only** if the adjacent **Vnet** has at least one **Vnic**
registered on the **Topic**.
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnet-multicast-cross-nodes.png)

## Unicast Topic <a name="unicasttopic"></a>

The **Vnic** can unicast a message to a **Topic**. The message will be delivered to **only one**
**Vnic** registered on the **Topic**. Unless explicitly specified, the message will be delivered via
the following fallback logic:

- Is sending **Vnic** registered on the **Topic**? Deliver to self. ->
- Explicit **Topic Leader** specified? Deliver to the **Topic Leader**. ->
- Is there a **Vnic** registered on the **Topic** in the same machine as the sender? Deliver to that **Vnic**. ->
- Deliver to the **Topic Leader**.

## Request/Reply <a name="requestreply"></a>

Request/Reply is essentially sending a message and waiting for the reply. It is utilizing
the Unicast & the Unicast Topic method is a synchronic way, expecting a reply message
from the target.

# Service Points - Standard API Sharing <a name="servicepoints"></a>

Project Home: https://github.com/saichler/servicepoints

## Overview

When a **Micro Service** is interacting with another **Micro Service**,
essentially it needs to invoke an API.
Using Client/Server technologies like Restful & GRPC isn't "ideal" (to say the least)
for internal Application communication. The other option is to use request/reply over
messaging system to invoke the internal API, however there are open challenges with Security
, AAA & Messages 2 API translation.

## Service Points

**Service Points** is encapsulating all the **Vnet Messaging, Security, AAA & the API** under a
simple interface that allows a transparent & seemless API invocation between one **Micro Service** to another, 
masking the networking interaction between services.

N number of services, each implemented as a service point, can reside inside the same **Process** or reside in
separated **Micro Services Processes**, all subject to the Author decision, The interaction is the same.
![alt text](https://github.com/saichler/layer8/blob/main/docs/service-points.png)

## Integrated Health Service & Leader/Follower Election <a name="health"></a>
The **Vnic/Vnet** is pre-integrated with health monitoring statistics **Service Point**. 
This service is monitoring the Memory & CPU of the hosting process, alongside a **Keep Alive**
heartbeat protocol.

The service is also integrated with **Leader/Follower** election.

## Invoking an API <a name="api"></a>

Invoking an API is simply utilizing the one of the GET, POST, PUT, PATCH, DELETE method on
the **Vnic**. The input is just the model instance and a **GSQL Query**
(https://github.com/saichler/gsql) in case of a GET. The Service Points framework
will encapsulate all the message interactions over the Vnet.
![alt text](https://github.com/saichler/layer8/blob/main/docs/api.png)

## Service Transactions <a name="transactions"></a>
A **Service Point** can be registered inside multiple **Vnic** (e.g. **Micro Services**), forming
A **Topic Overlay** for the provided service. In case of a **Stateful** service, the **Service Point**
has a michanism of a **Distributed Transaction** to apply a change or a state and rollback in case of
a failure in one of the instances.

## GSQL (Graph SQL) <a name="gsql"></a>

https://github.com/saichler/gsql

# From String <a name="fromstring"></a>

https://github.com/saichler/shared/tree/main/go/share/strings

# Introspector <a name="introspector"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/inspect

# Protobuf Object <a name="object"></a>

https://github.com/saichler/serializer/tree/main/go/serialize/object

# Meta Data Driven Property & Dynamic Instantiation <a name="property"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/property

# Updater & Generic Model Change Set <a name="updater"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/updater

# Deep Clone (Model Sensitive) <a name="clone"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/clone

# Distributed Cache & Delta Notifications <a name="cache"></a>

https://github.com/saichler/servicepoints/tree/main/go/points/cache

# Distributed Collection Service <a name="collect"></a>

https://github.com/saichler/collect/tree/main/go/collection/control

# Distributed Model Agnostic Parsing <a name="parse"></a>

https://github.com/saichler/collect/tree/main/go/collection/parsing

# Distributed Model Agnostic Inventory <a name="inventory"></a>

https://github.com/saichler/collect/tree/main/go/collection/inventory

# Traffic Generator <a name="generator"></a>

https://github.com/saichler/traffic

# Kubernetes Observer <a name="k8sobserve">

https://github.com/saichler/k8s_observer




