# ***Introducing Layer 8, Micro Service Layer into the OSI model***

# Overview
Roughly until 2013, an **Application** code was usually running inside a single process. 
Shared data & services between threads was usually done via a **Singleton** and the challenges with 
**Security**, **Concurrency** & **Scalability** within the **Application** 
were mainly with external API and integrations.
The traditional **OSI** model of the seven layers of networking, ends with the **Application layer**, 
which was **"satisfactory"** for the challenges of **Application** development.
![alt text](https://github.com/saichler/layer8/blob/main/osi.png)

Since 2013, **Containers** started to pick up, splitting, what once was a **single** process, **Application** 
into **Multiple Processes**, each fulfilling a logical part of the **Application** and 
**Servicing** internal functionality for the **Application**. 

### This is known as **Micro Services**. 
**Micro Services** need to work together, as one, to facade a single application to the user. 
As they need to exchange data, back and forth, between the processes to deliver the 
**Application** functionality, 
**Micro Services** use Networking to internally communicate and exchange this data, 
in other words, the era of **Internal Application Communication** has begun.

The new challenges, introduced with **Internal Application Communication**,
are large, expensive and hard to maintain, costing companies/projects
significant amount of **$$$$$** (educated guess is around hundred of millions per project), 
which is leading to an insight that there is a missing layer in the traditional **OSI Model**
that needs to standardize, simplify & secure those challenges, 
hence scientifically reducing the time & cost of developing a Micro Serives based application.
### This is the Layer 8, The Micro Services Layer
![alt text](https://github.com/saichler/layer8/blob/main/osi8.png)

# Please Explain?!
To simplify and emphasize the **Challenges**, will use an analogy to a **Person** and a **Job**.
 
```
Before: 
  A single Person used to do the Job.
```
```
After:
  With Micro Services, the Job is divided into several Tasks & Services,
  each is assigned to a different Person/s. Each Person sits in a different room.
```
Several people, each doing a part of the **Job** usually means they will complete the **Job** faster, 
enabling a bigger throughput of **Jobs** being done. ***However...
Basic prerequisite is needed before they reach this idle point...***

### The Team
A **Task/Service** of a **Job** has **owners**. 
**Owners** and not **Owner** as we have at least two, or more, people that can do that 
Task/Service for **High Availability** sake. **Arbitrary**, say there are 6 Tasks/Services, 
there will be at least **12 people** in the **Team**.

<sub>**Remember!** Each person is sitting in a different room, so we have 12 rooms.</sub>

### Initially
We need to establish an internal communication infrastructure so team members
will be able to phone and share information with each other.

### Rooms
Because rooms have different sizes and amenities, they are dynamically allocated to
a **Person** daily, and they don't have a landline with a fix number.

### Phones
When a **Person** is being assigned a room, he picks up a phone from the pool, 
start using the room, and return the phone at the end of the day. 
In other words, each day, the **Person** has a **new Phone number**.

## Challenge #1 - Phone Numbers
While the team can use phone numbers to communicate with each other, 
it's not practical. Phone numbers might change midday, and they will need to 
frequently go back and forth to the billboard to be updated with the latest list.
Another challenge is "Which phone number to use for the service?", we have at least two
**Persons** that can provide the service.

# <Work in progress, following is just some notes...>

# Layer 8 Solution

![alt text](https://github.com/saichler/layer8/blob/main/layer8.png)


## Challenges
It all starts with identifying the new **Micro Services** challenges. 
The following is some of the new challenges area and challenge questions, 
with the layer8 approach of solving those challenges.

### - Security
Multiple processes/containers means they need to communicate with each
other via the **Network**. Over the network immediatly raises the following:
- ***How do you make sure the communication is allowed?*** 

    A: ***Security Provider*** - Layer 8 provides a security provider interface, being utilized throught the components.
- ***How do you make sure the communication is secured?*** 

    A: Any data shared over the wire via layer8 is encrypted by the ***Security Provider***.
- ***Reading the code, I see there is only Shallow Security Provider?***
    
    A: Yes, Layer8 is abstracting security so anyone can provide the ***Security Provider***. 

### - Service Location/Address 
Communicating via the **Network** requires an address/host/ip (kind of like a phone number).
- ***How do you know the address?***
- ***How do you attend multiple addresses for the same service?***
- ***How do you push & handle this information between multiple services?***

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