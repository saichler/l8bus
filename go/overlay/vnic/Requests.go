package vnic

import "sync"

type Requests struct {
	pending map[int32]*Request
	mtx     *sync.Mutex
}

type Request struct {
	cond     *sync.Cond
	msgNum   int32
	response interface{}
}

func newRequests() *Requests {
	this := &Requests{}
	this.pending = make(map[int32]*Request)
	this.mtx = &sync.Mutex{}
	return this
}

func (this *Requests) newRequest(msgNum int32) *Request {
	request := &Request{}
	request.msgNum = msgNum
	request.cond = sync.NewCond(&sync.Mutex{})
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.pending[msgNum] = request
	return request
}

func (this *Requests) getRequest(msgNum int32) *Request {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	request := this.pending[msgNum]
	delete(this.pending, msgNum)
	return request
}
