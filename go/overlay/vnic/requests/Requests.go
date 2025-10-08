package requests

import (
	"sync"
	"time"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8utils/go/utils/strings"
)

type Requests struct {
	pending map[string]*Request
	mtx     *sync.Mutex
}

type Request struct {
	cond           *sync.Cond
	msgSource      string
	msgNum         uint32
	timeout        time.Duration
	timeoutReached bool
	response       ifs.IElements
	log            ifs.ILogger
}

func NewRequests() *Requests {
	this := &Requests{}
	this.pending = make(map[string]*Request)
	this.mtx = &sync.Mutex{}
	return this
}

func (this *Requests) NewRequest(msgNum uint32, msgSource string, timeoutInSeconds int, log ifs.ILogger) *Request {
	request := &Request{}
	request.msgNum = msgNum
	request.msgSource = msgSource
	request.timeout = time.Duration(timeoutInSeconds)
	request.cond = sync.NewCond(&sync.Mutex{})
	request.log = log

	key := requestKey(msgSource, msgNum)

	this.mtx.Lock()
	defer this.mtx.Unlock()
	_, ok := this.pending[key]
	if ok {
		panic("duplicated request")
	}
	this.pending[key] = request
	return request
}

func (this *Requests) GetRequest(msgNum uint32, msgSource string) *Request {
	key := requestKey(msgSource, msgNum)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	request := this.pending[key]
	return request
}

func (this *Requests) DelRequest(msgNum uint32, msgSource string) {
	key := requestKey(msgSource, msgNum)
	this.mtx.Lock()
	defer this.mtx.Unlock()
	delete(this.pending, key)
}

func (this *Request) Lock() {
	this.cond.L.Lock()
}

func (this *Request) Unlock() {
	this.cond.L.Unlock()
}

func (this *Request) MsgNum() uint32 {
	return this.msgNum
}

func (this *Request) MsgSource() string {
	return this.msgSource
}

func (this *Request) Response() ifs.IElements {
	if this.timeoutReached {
		return object.NewError("Timeout Reached!")
	}
	return this.response
}

func (this *Request) Wait() {
	go this.timeoutCheck()
	this.cond.Wait()
}

func (this *Request) timeoutCheck() {
	time.Sleep(time.Second * this.timeout)
	this.Lock()
	defer this.Unlock()
	if this.response == nil {
		this.timeoutReached = true
		this.cond.Broadcast()
	}
}

func (this *Request) SetResponse(resp ifs.IElements) {
	//The request timeout, so do nothing
	if this == nil {
		return
	}
	tr, ok := resp.Element().(*l8services.L8Transaction)
	this.response = resp
	if ok && tr.End == 0 {
		return
	}
	this.cond.Broadcast()
}

func requestKey(msgSource string, msgNum uint32) string {
	return strings.New(msgSource, int(msgNum)).String()
}
