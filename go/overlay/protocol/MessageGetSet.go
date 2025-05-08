package protocol

import (
	"github.com/saichler/l8types/go/ifs"
)

func (this *Message) Source() string {
	return string(this.source[0:])
}

func (this *Message) Vnet() string {
	return string(this.vnet[0:])
}

func (this *Message) Destination() string {
	return this.destination
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

func (this *Message) Priority() ifs.Priority {
	return this.priority
}

func (this *Message) Action() ifs.Action {
	return this.action
}

func (this *Message) SetAction(action ifs.Action) {
	this.action = action
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

func (this *Message) Tr() ifs.ITransaction {
	return this.tr
}

func (this *Message) SetTr(transaction ifs.ITransaction) {
	this.tr = transaction.(*Transaction)
}

func (this *Transaction) Id() string {
	return string(this.id[0:])
}

func (this *Transaction) State() ifs.TransactionState {
	return this.state
}

func (this *Transaction) SetState(st ifs.TransactionState) {
	this.state = st
}

func (this *Transaction) ErrorMessage() string {
	return this.errMsg
}

func (this *Transaction) SetErrorMessage(err string) {
	this.errMsg = err
}

func (this *Transaction) StartTime() int64 {
	return this.startTime
}
