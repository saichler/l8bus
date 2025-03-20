package protocol

import (
	"github.com/saichler/types/go/nets"
	"github.com/saichler/types/go/types"
)

const (
	UNICAST_ADDRESS_SIZE = 36
	INT_SIZE             = 4
	PRIORITY_SIZE        = 1
	HEADER_SIZE          = UNICAST_ADDRESS_SIZE*3 + INT_SIZE + PRIORITY_SIZE
	SOURCE_VNET_POS      = UNICAST_ADDRESS_SIZE
	DESTINATION_POS      = UNICAST_ADDRESS_SIZE * 2
	MULTICAST_POS        = UNICAST_ADDRESS_SIZE * 3
	VLAN_POS             = UNICAST_ADDRESS_SIZE * 4
	PRIORITY_POS         = UNICAST_ADDRESS_SIZE*4 + 4
)

func CreateHeader(msg *types.Message) []byte {
	header := make([]byte, HEADER_SIZE)
	for i, c := range msg.SourceUuid {
		header[i] = byte(c)
	}
	for i, c := range msg.SourceVnetUuid {
		header[i+SOURCE_VNET_POS] = byte(c)
	}
	for i, c := range msg.DestinationUuid {
		header[i+DESTINATION_POS] = byte(c)
	}
	for i, c := range msg.MulticastGroup {
		header[i+MULTICAST_POS] = byte(c)
	}
	vlan := nets.Int2Bytes(msg.Vlan)
	for i := 0; i < len(vlan); i++ {
		header[VLAN_POS+i] = vlan[i]
	}
	header[PRIORITY_POS] = byte(msg.Priority)
	return header
}

func stringOf(data []byte, start, end int) string {
	index := start
	for index = start; index < end; index++ {
		if data[index] == 0 {
			return string(data[start:index])
		}
	}
	return string(data[start:end])
}

func HeaderOf(data []byte) (string, string, string, string, int32, types.Priority) {
	// Source will always be Uuid
	source := string(data[0:UNICAST_ADDRESS_SIZE])
	// Source vnet will always be Uuid
	sourceVnet := string(data[SOURCE_VNET_POS:DESTINATION_POS])
	// destination will always be uuid, but may be missing
	destination := stringOf(data, DESTINATION_POS, MULTICAST_POS)
	// Multicast group, may be missing and not the full size of uuid
	multicast := stringOf(data, MULTICAST_POS, VLAN_POS)
	vlan := nets.Bytes2Int(data[VLAN_POS : VLAN_POS+4])
	pri := types.Priority(data[PRIORITY_POS])
	return source, sourceVnet, destination, multicast, vlan, pri
}
