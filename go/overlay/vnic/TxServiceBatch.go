package vnic

import (
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
)

type txServiceBatch struct {
	mtx         *sync.Mutex
	serviceName string
	serviceArea byte
	queue       []*txServiceBatchEntry
	interval    time.Duration
	mode        ifs.MulticastMode
	vnic        ifs.IVNic
}

type txServiceBatchEntry struct {
	element interface{}
	action  ifs.Action
}

func newTxServiceBatch(serviceName string, serviceArea byte, mode ifs.MulticastMode, interval int) *txServiceBatch {
	tsb := &txServiceBatch{}
	tsb.mtx = &sync.Mutex{}
	tsb.queue = make([]*txServiceBatchEntry, 0)
	tsb.serviceName = serviceName
	tsb.serviceArea = serviceArea
	tsb.interval = time.Duration(interval)
	return tsb
}

func (this *txServiceBatch) Send(action ifs.Action, element interface{}) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.queue = append(this.queue, &txServiceBatchEntry{element: element, action: action})
}

func (this *txServiceBatch) watch() {
	for this.vnic.Running() {
		this.flush()
		time.Sleep(time.Second * this.interval)
	}
}

func (this *txServiceBatch) flush() {
	this.mtx.Lock()
	items := this.queue
	this.queue = make([]*txServiceBatchEntry, 0)
	defer this.mtx.Unlock()
	if len(this.queue) > 0 {
		var list []interface{}
		lastAction := -1
		for _, item := range items {
			if lastAction != int(item.action) {
				if list != nil {
					this.send(ifs.Action(lastAction), list)
				}
				list = make([]interface{}, 0)
				lastAction = int(item.action)
			}
			list = append(list, item.element)
		}
		if list != nil {
			this.send(ifs.Action(lastAction), list)
		}
	}
}

func (this *txServiceBatch) send(action ifs.Action, elements []interface{}) {
	switch this.mode {
	case ifs.M_Leader:
		this.vnic.Leader(this.serviceName, this.serviceArea, action, elements)
	case ifs.M_Proximity:
		this.vnic.Proximity(this.serviceName, this.serviceArea, action, elements)
	case ifs.M_RoundRobin:
		this.vnic.RoundRobin(this.serviceName, this.serviceArea, action, elements)
	case ifs.M_Local:
		this.vnic.Local(this.serviceName, this.serviceArea, action, elements)
	case ifs.M_All:
		this.vnic.Multicast(this.serviceName, this.serviceArea, action, elements)
	default:
		panic("Unsupported mode")
	}
}

func BufferKey(serviceName string, serviceArea byte, mode ifs.MulticastMode) string {
	return strings.New(serviceName, serviceArea, mode).String()
}
