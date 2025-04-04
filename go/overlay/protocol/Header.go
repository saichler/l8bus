package protocol

/*
import (
	"github.com/saichler/types/go/nets"
	"github.com/saichler/types/go/types"
)

const (
	UNICAST_ADDRESS_SIZE = 36
	INT_SIZE             = 4
	PRIORITY_SIZE        = 1
	SOURCE_POS           = 0
	VNET_POS             = UNICAST_ADDRESS_SIZE
	DESTINATION_POS      = VNET_POS + UNICAST_ADDRESS_SIZE
	SERVICE_AREA_POS     = DESTINATION_POS + 2
	SERVICE_NAME_POS     = DESTINATION_POS + UNICAST_ADDRESS_SIZE
	PRIORITY_POS         = SERVICE_AREA_POS + 4
	HEADER_SIZE          = PRIORITY_POS + 1
)

func CreateHeader(msg *types.Message) []byte {
	header := make([]byte, HEADER_SIZE)
	for i, c := range msg.Source {
		header[i] = byte(c)
	}
	for i, c := range msg.SourceVnet {
		header[i+SOURCE_VNET_POS] = byte(c)
	}
	for i, c := range msg.Destination {
		header[i+DESTINATION_POS] = byte(c)
	}
	for i, c := range msg.ServiceName {
		header[i+SERVICE_NAME_POS] = byte(c)
	}
	serviceArea := nets.Int2Bytes(msg.ServiceArea)
	for i := 0; i < len(serviceArea); i++ {
		header[SERVICE_AREA_POS+i] = serviceArea[i]
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
	destination := stringOf(data, DESTINATION_POS, SERVICE_NAME_POS)
	// Multicast group, may be missing and not the full size of uuid
	serviceName := stringOf(data, SERVICE_NAME_POS, SERVICE_AREA_POS)
	serviceArea := nets.Bytes2Int(data[SERVICE_AREA_POS : SERVICE_AREA_POS+4])
	pri := types.Priority(data[PRIORITY_POS])
	return source, sourceVnet, destination, serviceName, serviceArea, pri
}
*/
