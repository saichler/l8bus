package switching

import (
	logs "github.com/saichler/shared/go/share/interfaces"
	"net"
	"strings"
)

var ipSegment = newIpAddressSegment()

// IPSegment Let the switching know if the incoming ip belongs to this machine/vm or is it external machine/vm.
type IPSegment struct {
	ip2IfName    map[string]string
	subnet2Local map[string]bool
}

// Initialize
func newIpAddressSegment() *IPSegment {
	ias := &IPSegment{}
	lip, err := localIps()
	if err != nil {
		logs.Error(err)
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
			ias.subnet2Local[subnet(ip)] = true
		} else if name[0:3] == "eth" || name[0:3] == "ens" {
			ias.subnet2Local[subnet(ip)] = false
		} else {
			ias.subnet2Local[subnet(ip)] = true
		}
	}
}

// Check if this ip's subnet is within the local subnet list
func (ias *IPSegment) isLocal(ip string) bool {
	return ias.subnet2Local[subnet(ip)]
}

// look for the subnet facing public networking, e.g. the ip on eth0 & etc.
// @TODO - Add support for multiple NICs
func (ias *IPSegment) externalSubnet() string {
	for subnet, isLocal := range ias.subnet2Local {
		if !isLocal {
			return subnet
		}
	}
	return "No External Subnet"
}

// substr the subnet from an ip
// @TODO - add support for ipv6
func subnet(ip string) string {
	index2 := strings.LastIndex(ip, ".")
	if index2 != -1 {
		return ip[0:index2]
	}
	return ip
}

// Iterate over the machine interfaces and map the ip to the interface name
func localIps() (map[string]string, error) {
	netIfs, err := net.Interfaces()
	if err != nil {
		return nil, logs.Error("Could not fetch local interfaces:", err.Error())
	}
	result := make(map[string]string)
	for _, netIf := range netIfs {
		addrs, err := netIf.Addrs()
		if err != nil {
			logs.Error("Failed to fetch addresses for net interface:", err.Error())
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
