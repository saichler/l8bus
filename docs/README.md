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

01. [Security Provider](#security)
02. [Vnet (Virtual Network)](#vnet)
03. [Vnic (Virtual Network Interface)](#Vnic)
04. [Unicast](#unicast)
05. [Multicast](#multicast)
06. [Unicast Topic](#unicasttopic)
07. [Request/Reply](#requestreply)
08. [Service Points](#servicepoints)
09. [Invoking an API](#api)
10. [GSQL (Graph SQL)](#gsql)
11. [From String](#fromstring)
12. [Introspector](#introspectpr)
13. [Protobug Object](#object)
14. [Meta Data Driven Property & Dynamic Instantiation](#property)
15. [Updater & Generic Model Change Set](#updater)
16. [Deep Clone (Model Sensitive)](#clone)
17. [Distributed Cache & Delta Notifications](#cache)
18. [Distributed Collection Service](#collect)
19. [Distributed Model Agnostic Parsing](#parse)
20. [Distributed Model Agnostic Inventory](#inventory)
21. [Traffic Generator](#generator)

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
simple interface that allows a transparent & seemless API invocation between one **Micro Service** to another.
![alt text](https://github.com/saichler/layer8/blob/main/docs/service-points.png)

## Invoking an API <a name="api"></a>

Invoking an API is simply utilizing the one of the GET, POST, PUT, PATCH, DELETE method on
the **Vnic**. The input is just the model instance and a **GSQL Query**
(https://github.com/saichler/gsql) in case of a GET. The Service Points framework
will encapsulate all the message interactions over the Vnet.
![alt text](https://github.com/saichler/layer8/blob/main/docs/api.png)

## GSQL (Graph SQL) <a name="gsql"></a>

https://github.com/saichler/gsql

# From String <a name="fromstring"></a>

https://github.com/saichler/shared/tree/main/go/share/strings

# Introspector <a name="vnic">introspector</a>

https://github.com/saichler/reflect/tree/main/go/reflect/inspect

# Protobuf Object <a name="object"></a>

https://github.com/saichler/serializer/tree/main/go/serialize/object

# Meta Data Driven Property & Dynamic Instantiation <a name="property"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/property

# Updater & Generic Model Change Set <a name="updater"></a>

https://github.com/saichler/reflect/tree/main/go/reflect/updater

# Deep Clone (Model Sensitive) <a name="deepclone"></a>

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




