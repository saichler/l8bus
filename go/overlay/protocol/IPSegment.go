package protocol

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
)

var IpSegment = newIpAddressSegment()
var UsingContainers = true
var MachineIP = "127.0.0.1"

// IPSegment Let the switching know if the incoming ip belongs to this machine/vm or is it external machine/vm.
type IPSegment struct {
	ip2IfName    map[string]string
	subnet2Local map[string]bool
}

// Initialize
func newIpAddressSegment() *IPSegment {
	ias := &IPSegment{}
	lip, err := LocalIps()
	if err != nil {
		panic(err)
	}
	ias.ip2IfName = lip
	ias.initSegment()
	return ias
}

// Initiate and destinguish all the interfaces if they are local or public
// @TODO - Find a more elegant way to determinate this, like a map
func (ias *IPSegment) initSegment() {
	ias.subnet2Local = make(map[string]bool)
	for ip, name := range ias.ip2IfName {
		if name == "lo" {
			ias.subnet2Local[Subnet(ip)] = true
		} else if name[0:3] == "eth" ||
			name[0:3] == "ens" ||
			name[0:3] == "en0" {
			ias.subnet2Local[Subnet(ip)] = false
			if MachineIP == "127.0.0.1" {
				MachineIP = ip
			} else if strings.Contains(MachineIP, ":") {
				MachineIP = ip
			}
		} else {
			ias.subnet2Local[Subnet(ip)] = true
		}
	}
}

// Check if this ip's subnet is within the local subnet list
func (ias *IPSegment) IsLocal(ip string) bool {
	ip = IP(ip)
	if ip == MachineIP {
		return true
	}
	return ias.subnet2Local[Subnet(ip)]
}

// look for the subnet facing public networking, e.g. the ip on eth0 & etc.
// @TODO - Add support for multiple NICs
func (ias *IPSegment) ExternalSubnet() string {
	for subnet, isLocal := range ias.subnet2Local {
		if !isLocal {
			return subnet
		}
	}
	return ""
}

// substr the subnet from an ip
// @TODO - add support for ipv6
func Subnet(ip string) string {
	index2 := strings.LastIndex(ip, ".")
	if index2 != -1 {
		return ip[0:index2]
	}
	return ip
}

func IP(ip string) string {
	index := strings.Index(ip, "/")
	if index != -1 {
		return ip[0:index]
	}
	index = strings.LastIndex(ip, ":")
	if index != -1 {
		return ip[0:index]
	}
	return ip
}

// Iterate over the machine interfaces and map the ip to the interface name
func LocalIps() (map[string]string, error) {
	fmt.Println("GOOS=", runtime.GOOS)
	
	netIfs, err := net.Interfaces()
	if err != nil {
		return nil, errors.New("Could not fetch local interfaces: " + err.Error())
	}
	result := make(map[string]string)
	for _, netIf := range netIfs {
		addrs, err := netIf.Addrs()
		if err != nil {
			//logs.Error("Failed to fetch addresses for net interface:", err.Error())
			continue
		}
		for _, addr := range addrs {
			addrString := addr.String()
			index := strings.Index(addrString, "/")
			result[addrString[0:index]] = netIf.Name
		}
	}
	return result, nil
}
