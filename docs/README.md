# Layer 8 Functionality

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

## Security Provider
When developing a Micro Services stack, usually security consideration comes a bit late
in the game, creating tough challenges. With **Layer8**, this is the starting point!

The **Security Provider** is an abstraction inteface for AAA & Encryption, being used and utilized
the development stack & frameworks. **The Prime rule** is that two component need to have
the **same Security Provider** to interact with each other.

What is the **Same Security Provider?** As this is an interfaced abstraction,
the implementation is currently private...;o)

## Vnet (Virtual Network)
**Vnet** is a process running on each host as an OS service, 
that we want as part of the **Application Overlay**. 
There is the flexability to have multiple Application Vnet hosted inside the same OS service
or an OS service per Application Vnet
![alt text](https://github.com/saichler/layer8/blob/main/docs/vnet.png)

## Vnic
The **VNic** is a piece of code/library used inside the running process to connect and send/publish/request messages
inside the **Vnet**. When instantiated, it autodetect and connects to the Vnet, **given**
it has the correct **Security Provider**. It is **agnostic** to being hosted inside K8s, Docker, Container or plain process. 
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnic2vnet-connect.png)

## Unicast sequence
The Vnic can unicast a message to another Vnic on the Vnet via its uuid address.
### Unicast flow
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnet-unicast-cross-nodes.png)

## Multicast sequence
A Vnic can publish a message to a topic over the Vnet. The Vnet agents will receive the multicast
only if they have at least 1 Vnic that is registered to consume the topic.
### Multicast flow
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnet-multicast-cross-nodes.png)



