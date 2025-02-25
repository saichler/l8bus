# Layer 8 Functionality

## Security Provider Plugin
The Security Provider Plugin is an abstraction for AAA & Encryption, being used and
invoked at different points of the framework, authentication users, validating service connections, 
API access & data scopes.

## Vnet & Vnic
The **Vnet** is a service running on the host. It is using the Security Provider Plugin to validate
internal & external connection requests from Vnics. Once connected, the **Vnic** info is
added to the switching table of the **Vnet**.

The **VNic** is a piece of code/library used inside the provess to connect and send/public messages
inside the **Vnet**.

### Vnic connects to the Vnet sequence
![alt text](https://github.com/saichler/layer8/blob/main/docs/layer-8-vnic2vnet-connect.png)

