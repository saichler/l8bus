package vnet

import (
	"github.com/saichler/layer8/go/overlay/protocol"
	"net"
	"strconv"
	"strings"
	"time"
)

func (this *VNet) Discover() {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(this.resources.SysConfig().VnetPort-2)))
	if err != nil {
		this.resources.Logger().Error("Discovery: ", err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		this.resources.Logger().Error("Discovery: ", err.Error())
		return
	}
	this.udp = conn
	go this.discoveryRx()
	go this.Broadcast()
}

func (this *VNet) discoveryRx() {
	packet := []byte{0, 0, 0}
	defer this.udp.Close()

	for this.running {
		n, addr, err := this.udp.ReadFromUDP(packet)
		this.resources.Logger().Debug("Recevied discovery broadcast")
		if !this.running {
			break
		}
		if err != nil {
			this.resources.Logger().Error(err.Error())
			break
		}
		if n == 3 {
			ip := addr.IP.String()
			if ip != protocol.MachineIP && ip != "127.0.0.1" {
				if strings.Compare(ip, protocol.MachineIP) == -1 && !this.switchTable.conns.isConnected(ip) {
					this.resources.Logger().Info("Trying to connect to peer at ", ip)
					err = this.ConnectNetworks(ip, this.resources.SysConfig().VnetPort)
					if err != nil {
						this.resources.Logger().Error("Discovery: ", err.Error())
					}
				}
			}
		}
	}
}

func (this *VNet) Broadcast() {
	this.resources.Logger().Debug("Sending discovery broadcast")
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(int(this.resources.SysConfig().VnetPort-2)))
	if err != nil {
		this.resources.Logger().Error("Failed to resolve broadcast:", err.Error())
		return
	}
	this.udp.WriteToUDP([]byte{1, 2, 3}, addr)
	time.Sleep(time.Second * 10)
	this.resources.Logger().Debug("Sending discovery broadcast")
	this.udp.WriteToUDP([]byte{1, 2, 3}, addr)
	for this.running {
		time.Sleep(time.Minute)
		this.resources.Logger().Debug("Sending discovery broadcast")
		this.udp.WriteToUDP([]byte{1, 2, 3}, addr)
	}
}
