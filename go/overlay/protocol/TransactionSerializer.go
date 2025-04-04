package protocol

import (
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/nets"
)

const (
	POS_Tr_Id          = 1
	POS_Tr_State       = POS_Tr_Id + 36
	POS_Tr_Start_Time  = POS_Tr_State + 1
	POS_Tr_Err_Message = POS_Tr_Start_Time + 8
)

type TransactionSerializer struct {
}

func (this *TransactionSerializer) Mode() common.SerializerMode {
	return common.BINARY
}

func (this *TransactionSerializer) Marshal(any interface{}, r common.IRegistry) ([]byte, error) {
	if common.IsNil(any) {
		return []byte{0}, nil
	}
	tr := any.(*Transaction)
	POS_END := POS_Tr_Err_Message + 3 + len(tr.errMsg)
	data := make([]byte, POS_END)
	data[0] = 1

	copy(data[POS_Tr_Id:POS_Tr_State], tr.id[0:36])
	data[POS_Tr_State] = byte(tr.state)
	copy(data[POS_Tr_Start_Time:POS_Tr_Err_Message], nets.Long2Bytes(tr.startTime))
	copy(data[POS_Tr_Err_Message:POS_Tr_Err_Message+2], nets.UInt162Bytes(uint16(len(tr.errMsg))))
	copy(data[POS_Tr_Err_Message+2:POS_END], tr.errMsg)
	return data, nil
}

func (this *TransactionSerializer) Unmarshal(data []byte, r common.IRegistry) (interface{}, error) {
	if len(data) == 1 {
		return nil, nil
	}
	tr := &Transaction{}
	copy(tr.id[0:36], data[POS_Tr_Id:POS_Tr_State])
	tr.state = common.TransactionState(data[POS_Tr_State])
	tr.startTime = nets.Bytes2Long(data[POS_Tr_Start_Time:POS_Tr_Err_Message])
	size := nets.Bytes2UInt16(data[POS_Tr_Err_Message : POS_Tr_Err_Message+2])
	POS_END := POS_Tr_Err_Message + 2 + int(size)
	tr.errMsg = string(data[POS_Tr_Err_Message+2 : POS_END])
	return tr, nil
}
