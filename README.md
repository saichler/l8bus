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
hence significantly reducing the time & cost of developing a Micro Serives based application.
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
In other words, each day, the **Person** might have a **new Phone number**.

## Challenge #1 - Phone Numbers (IP Addresses)
While the team can use phone numbers to communicate with each other, 
it's not practical. Phone numbers keep changing... **They will need to set up 
some system to frequently sync/notify phone number changes.**

## Challenge #2 - Hire Coordinators (CNI, Core DNS & Kube Proxy)
To overcome challenge #1, we need to hire 3 more people:
- **A Phone manager (CNI)**
  Hand, and collect the phone from each **Person**, making sure they work.
- **A Service to Phones coordinator (DNS)**
  Keep track of which phone numbers provide which service so a Person can call
  the DNS and request a phone for a service.
- **A Service to Rooms coordinator (Kube Proxy)**
  In case People need to share documents, which room currently belongs to which
  **Person**.

# Challenge #3 - Internal Integration
Now that the communication layer between the **Team Members** (OSI model Layer 7 - Application) 
was established, we reach the end of the **"Guided" OSI Model**, and left with the greatest challenge
of them all, **Internal Integration**.

**Internal Integration** occurs when two team members, or more, are brought together to collaborate
and provide a service to external consumers/customers. 
With the missing OSI Layer 8, **Internal Integration** is treated as **Integration**, 
which is **the most painful, time-consuming, heavy maintenance, money consuming element in the software
lifecycle**.

## Integration
When a **Person** integrating with another **Person**, they need to establish and agree on
the way they interact with each other, this is called **API** (Application Programming Interface). 
API includes:
- **Words (Protocol & Serialization)**
While both know english, They need to agree on **words** that derive the conversation. 
For example, every interaction starts with "Hello", following with "Please do/Please get/Please update/...". 
This is not staingth forward for the un-experienced, some choose written
language instead of spoken language...

- **Subjects (Data Types)**
What are the subjects that **Person A** is knowledgeable about that **Person B** needs to do his task?
- **Work Method (StateLess/Stateful)**
- **Work Load (Scalability)**
- **Work Hours (Hi-Availability)**
- **Authentication, Authorization & Accounting (Security)**

## Insight 1
The above is quantified to a **huge amount** of time, effort & money that is **repetitively done** between each two persons of the **same team**.

## Insight 2
**Each one** of the steps above contains a **Deep Pothole**. 
Falling, at any step, to one of those **Potholes** is **Magnitude** the cost of **Insight 1**.

# Exiting the analogy and Back to Micro Services.


# <Work in progress, following is just some notes...>

# Layer 8 Solution

![alt text](https://github.com/saichler/layer8/blob/main/layer8.png)
