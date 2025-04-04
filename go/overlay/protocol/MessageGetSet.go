package protocol

import "github.com/saichler/types/go/common"

func (this *Message) Source() string {
	return string(this.source[0:])
}

func (this *Message) Vnet() string {
	return string(this.vnet[0:])
}

func (this *Message) Destination() string {
	return string(this.destination[0:])
}

func (this *Message) ServiceName() string {
	return this.serviceName
}

func (this *Message) ServiceArea() uint16 {
	return this.serviceArea
}

func (this *Message) Sequence() uint32 {
	return this.sequence
}

func (this *Message) Priority() common.Priority {
	return this.priority
}

func (this *Message) Action() common.Action {
	return this.action
}

func (this *Message) Timeout() uint16 {
	return this.timeout
}

func (this *Message) Request() bool {
	return this.request
}

func (this *Message) Reply() bool {
	return this.reply
}

func (this *Message) FailMessage() string {
	return this.failMessage
}

func (this *Message) Data() string {
	return this.data
}

func (this *Message) SetData(data string) {
	this.data = data
}

func (this *Message) Tr() common.ITransaction {
	return this.tr
}

func (this *Transaction) Id() string {
	return string(this.id[0:])
}

func (this *Transaction) State() common.TransactionState {
	return this.state
}

func (this *Transaction) ErrorMessage() string {
	return this.errMsg
}

func (this *Transaction) StartTime() int64 {
	return this.startTime
}
