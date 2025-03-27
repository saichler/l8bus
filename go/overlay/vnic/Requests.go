package vnic

import (
	"bytes"
	"github.com/saichler/types/go/common"
	"strconv"
	"sync"
)

type Requests struct {
	pending map[string]*Request
	mtx     *sync.Mutex
}

type Request struct {
	cond      *sync.Cond
	msgSource string
	msgNum    int32
	response  common.IMObjects
}

func newRequests() *Requests {
	this := &Requests{}
	this.pending = make(map[string]*Request)
	this.mtx = &sync.Mutex{}
	return this
}

func (this *Requests) newRequest(msgNum int32, msgSource string) *Request {
	request := &Request{}
	request.msgNum = msgNum
	request.msgSource = msgSource
	request.cond = sync.NewCond(&sync.Mutex{})
	key := bytes.Buffer{}
	key.WriteString(msgSource)
	key.WriteString(strconv.Itoa(int(msgNum)))
	this.mtx.Lock()
	defer this.mtx.Unlock()
	_, ok := this.pending[key.String()]
	if ok {
		panic("duplicated request")
	}
	this.pending[key.String()] = request
	return request
}

func (this *Requests) getRequest(msgNum int32, msgSource string) *Request {
	key := bytes.Buffer{}
	key.WriteString(msgSource)
	key.WriteString(strconv.Itoa(int(msgNum)))
	this.mtx.Lock()
	defer this.mtx.Unlock()
	request := this.pending[key.String()]
	delete(this.pending, key.String())
	return request
}
