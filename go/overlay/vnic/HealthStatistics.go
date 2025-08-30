package vnic

import (
	"sync/atomic"
	"time"
)

type HealthStatistics struct {
	LastMsgTime atomic.Int64
	TxMsgCount  atomic.Int64
	TxDataCount atomic.Int64
	RxMsgCount  atomic.Int64
	RxDataCont  atomic.Int64
}

func (this *HealthStatistics) Stamp() {
	this.LastMsgTime.Store(time.Now().UnixMilli())
}

func (this *HealthStatistics) IncrementTX(data []byte) {
	this.TxMsgCount.Add(1)
	this.TxDataCount.Add(int64(len(data)))
}

func (this *HealthStatistics) IncrementRx(data []byte) {
	this.RxMsgCount.Add(1)
	this.RxDataCont.Add(int64(len(data)))
}
