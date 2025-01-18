# **Work In Progress**

# Layer8
### Process to Process data share made secure & easy.
# Prolog
Until the early 21st century, developing an application ment code was expected to run in a single process. 
Shared data & services between threads was usually done via a **Singleton** and the challenges with  
**Security** & **Concurrency** were mainly with external API and integrations.

Roughly around 2013, **Containers** started to pick up with the concept of **Micro Services**. 
The Micro Services concept is splitting the code base, once running in a **single** process, into 
**multiple processes deployed on multiple machines**.
In other words, software was **"broken"** into small pices, each running in its own process/container and distributed on multiple machines.

While **Micro Servicing** software derives abstraction, scalability and easier maintainability, it presents new 
challenges that did not exist before... Data location, sharing, concurrency & security became a common
infrastructure challenge as now processes need to share and exchange data **via the Network**. 
  

## Overview
The new challenges of **Micro Services** were not met with a standard infrastructure, 
instead each adapting project, using software design for the 
**single process era, re-inventing the wheel...** The outcome 
is a **Billion $$$** of developing & maintaining software infrastructure, 
again and again, for each project.

**Layer8** in trying to encapsulate the **Micro Services** infrastructure challenges into 
an agnostic, simple & maintainable framework, so developing **Micro Services** base application 
becomes seemless, easy & expedited.

![alt text](https://github.com/saichler/layer8/blob/main/layer8.png)

## Challenges
It all starts with identifying the new **Micro Services** challenges:
### - Service Location/Address
Multiple processes/containers means they need to communicate with each 
other via the **Network**. 
Communicating via the **Network** requires an address/host/ip (kind of like a phone number). 

- ***How do you know the address?***
- ***How do you attend multiple addresses for the same service?***

How do you push & handle this information between multiple services?

### - Service API
Networking is basically sending/receiving bytes. 
- ***What is the meaning of those bytes?***
- ***What action should the service take?***
- ***What is the response the service should reply?***
- ***What is the meaning of the reply?***

This effort is done between each two processes...

**The effort is equivalent to inventing the words & sentences for a new Laungage...**

### - Serialization
Processes use models/structure to represent their data models. 
When communicating between processes over the network, 
translation from the data to a set of bytes and back is needed.
- ***How do you do this translation efficiently?***
- ***When only a subset of the model/struct has changed, 
how do you just send and consume those changes seemlesly?***

### - High Availability
Dealing with multiple instances of the service, we have some challenges with high availability.
- ***When one of the instances is down, 
how do we shift the load to a new instance?***
- ***Data concurrency, how do we promise that with multiple instances?***

### - Scalability
### - Stateful & Stateless
### - Security