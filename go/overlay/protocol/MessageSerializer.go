package protocol

import (
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/nets"
)

var tser = &TransactionSerializer{}

const (
	ADDR_SIZE        = 36
	POS_SOURCE       = 0
	POS_VNET         = ADDR_SIZE
	POS_DESTINATION  = POS_VNET + ADDR_SIZE
	POS_SERVICE_AREA = POS_DESTINATION + ADDR_SIZE
	POS_SERVICE_NAME = POS_SERVICE_AREA + 2

	POS_SEQUENCE      = 0
	POS_PRIORITY      = POS_SEQUENCE + 4
	POS_ACTION        = POS_PRIORITY + 1
	POS_TIMEOUT       = POS_ACTION + 1
	POS_REQUEST_REPLY = POS_TIMEOUT + 2
	POS_AAAID         = POS_REQUEST_REPLY + 1
)

type MessageSerializer struct {
}

func (this *MessageSerializer) Mode() ifs.SerializerMode {
	return ifs.BINARY
}

func createHeader(message *Message) []byte {
	headerEnd := POS_SERVICE_NAME + len(message.serviceName) + 1
	header := make([]byte, headerEnd)
	copy(header[POS_SOURCE:POS_VNET], message.source[0:ADDR_SIZE])
	copy(header[POS_VNET:POS_DESTINATION], message.vnet[0:ADDR_SIZE])
	copy(header[POS_DESTINATION:POS_SERVICE_AREA], message.destination[0:ADDR_SIZE])
	copy(header[POS_SERVICE_AREA:POS_SERVICE_NAME], nets.UInt162Bytes(message.serviceArea))
	header[POS_SERVICE_NAME] = byte(len(message.serviceName))
	copy(header[POS_SERVICE_NAME+1:headerEnd], message.serviceName)
	return header
}

func createBody(message *Message, s ifs.ISecurityProvider) ([]byte, error) {

	trData, _ := tser.Marshal(message.Tr(), nil)

	posFailMessage := POS_AAAID + len(message.aaaId) + 1
	posData := posFailMessage + len(message.failMessage) + 1
	posTr := posData + len(message.data) + 4
	posEnd := posTr + len(trData)

	body := make([]byte, posEnd)

	copy(body[POS_SEQUENCE:POS_PRIORITY], nets.UInt322Bytes(message.sequence))
	body[POS_PRIORITY] = byte(message.priority)
	body[POS_ACTION] = byte(message.action)
	copy(body[POS_TIMEOUT:POS_REQUEST_REPLY], nets.UInt162Bytes(message.timeout))
	body[POS_REQUEST_REPLY] = nets.ByteOf(message.request, message.reply)
	body[POS_AAAID] = byte(len(message.aaaId))
	copy(body[POS_AAAID+1:posFailMessage], message.aaaId)
	body[posFailMessage] = byte(len(message.failMessage))
	copy(body[posFailMessage+1:posData], message.failMessage)
	copy(body[posData:posData+4], nets.UInt322Bytes(uint32(len(message.data))))
	copy(body[posData+4:posTr], message.data)
	copy(body[posTr:posEnd], trData)

	data, err := s.Encrypt(body)
	return []byte(data), err
}

func (this *MessageSerializer) Marshal(any interface{}, resources ifs.IResources) ([]byte, error) {
	message := any.(*Message)
	header := createHeader(message)
	body, err := createBody(message, resources.Security())
	if err != nil {
		return nil, err
	}
	data := make([]byte, len(header)+len(body))
	copy(data[0:len(header)], header)
	copy(data[len(header):], body)
	return data, nil
}

func HeaderOf(data []byte) (string, string, string, string, uint16) {
	source := string(data[POS_SOURCE:ADDR_SIZE])
	vnet := string(data[POS_VNET:POS_DESTINATION])
	destination := ""
	if data[POS_DESTINATION] != 0 && data[POS_DESTINATION+1] != 0 {
		destination = string(data[POS_DESTINATION:POS_SERVICE_AREA])
	}
	serviceArea := nets.Bytes2UInt16(data[POS_SERVICE_AREA:POS_SERVICE_NAME])
	serviceName := string(data[POS_SERVICE_NAME+1 : POS_SERVICE_NAME+1+int(data[POS_SERVICE_NAME])])
	return source, vnet, destination, serviceName, serviceArea
}

func populateHeader(data []byte, message *Message) {
	copy(message.source[0:], data[POS_SOURCE:POS_VNET])
	copy(message.vnet[0:], data[POS_VNET:POS_DESTINATION])
	copy(message.destination[0:], data[POS_DESTINATION:POS_SERVICE_AREA])
	message.serviceArea = nets.Bytes2UInt16(data[POS_SERVICE_AREA:POS_SERVICE_NAME])
	message.serviceName = string(data[POS_SERVICE_NAME+1 : POS_SERVICE_NAME+1+int(data[POS_SERVICE_NAME])])
}

func populateBody(d []byte, message *Message, s ifs.ISecurityProvider) error {
	headerSize := POS_SERVICE_NAME + 1 + int(d[POS_SERVICE_NAME])
	encData := d[headerSize:]
	body, err := s.Decrypt(string(encData))
	if err != nil {
		return err
	}

	message.sequence = nets.Bytes2UInt32(body[POS_SEQUENCE:POS_PRIORITY])
	message.priority = ifs.Priority(body[POS_PRIORITY])
	message.action = ifs.Action(body[POS_ACTION])
	message.timeout = nets.Bytes2UInt16(body[POS_TIMEOUT:POS_REQUEST_REPLY])
	message.request, message.reply = nets.BoolOf(body[POS_REQUEST_REPLY])
	posFailMessage := POS_AAAID + 1 + int(body[POS_AAAID])
	message.aaaId = string(body[POS_AAAID+1 : posFailMessage])
	posData := posFailMessage + 1 + int(body[posFailMessage])
	message.failMessage = string(body[posFailMessage+1 : posData])
	posTr := posData + 4 + int(nets.Bytes2UInt32(body[posData:posData+4]))
	message.data = string(body[posData+4 : posTr])
	tr, _ := tser.Unmarshal(body[posTr:], nil)
	if tr != nil {
		message.tr = tr.(*Transaction)
	}
	return nil
}

func (this *MessageSerializer) Unmarshal(data []byte, resources ifs.IResources) (interface{}, error) {
	msg := &Message{}
	populateHeader(data, msg)
	err := populateBody(data, msg, resources.Security())
	return msg, err
}
