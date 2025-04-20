package vnet

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"net"
	"strconv"
	"strings"
	"time"
)

type Discovery struct {
	vnet       *VNet
	conn       *net.UDPConn
	discovered map[string]bool
}

func NewDiscovery(vnet *VNet) *Discovery {
	ds := &Discovery{}
	ds.vnet = vnet
	ds.discovered = make(map[string]bool)
	return ds
}

func (this *Discovery) Discover() {
	if protocol.MachineIP == "127.0.0.1" {
		this.vnet.resources.Logger().Info("Discovery is disabled, machine IP is ", protocol.MachineIP)
		return
	}
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(this.vnet.resources.SysConfig().VnetPort-2)))
	if err != nil {
		this.vnet.resources.Logger().Error("Discovery: ", err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		this.vnet.resources.Logger().Error("Discovery: ", err.Error())
		return
	}
	this.conn = conn
	go this.discoveryRx()
	go this.Broadcast()
}

func (this *Discovery) discoveryRx() {
	this.vnet.resources.Logger().Debug("Listening for discovery broadcast")
	packet := []byte{0, 0, 0}
	defer this.conn.Close()

	for this.vnet.running {
		n, addr, err := this.conn.ReadFromUDP(packet)
		ip := addr.IP.String()
		this.vnet.resources.Logger().Debug("Recevied discovery broadcast from ", ip, " size ", n)
		if !this.vnet.running {
			break
		}
		if err != nil {
			this.vnet.resources.Logger().Error(err.Error())
			break
		}
		if n == 3 {
			if ip != protocol.MachineIP && ip != "127.0.0.1" {
				_, ok := this.discovered[ip]
				if strings.Compare(ip, protocol.MachineIP) == -1 && !ok {
					this.vnet.resources.Logger().Info("Trying to connect to peer at ", ip)
					err = this.vnet.ConnectNetworks(ip, this.vnet.resources.SysConfig().VnetPort)
					if err != nil {
						this.vnet.resources.Logger().Error("Discovery: ", err.Error())
					}
				}
				this.discovered[ip] = true
			}
		}
	}
}

func (this *Discovery) Broadcast() {
	this.vnet.resources.Logger().Debug("Sending discovery broadcast")
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+
		strconv.Itoa(int(this.vnet.resources.SysConfig().VnetPort-2)))
	if err != nil {
		this.vnet.resources.Logger().Error("Failed to resolve broadcast:", err.Error())
		return
	}
	this.conn.WriteToUDP([]byte{1, 2, 3}, addr)
	time.Sleep(time.Second * 10)
	this.vnet.resources.Logger().Debug("Sending discovery broadcast")
	this.conn.WriteToUDP([]byte{1, 2, 3}, addr)
	for this.vnet.running {
		time.Sleep(time.Minute)
		this.vnet.resources.Logger().Debug("Sending discovery broadcast")
		this.conn.WriteToUDP([]byte{1, 2, 3}, addr)
	}
}
