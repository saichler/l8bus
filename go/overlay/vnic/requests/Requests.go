package requests

import (
	"bytes"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"strconv"
	"sync"
	"time"
)

type Requests struct {
	pending map[string]*Request
	mtx     *sync.Mutex
}

type Request struct {
	cond           *sync.Cond
	msgSource      string
	msgNum         uint32
	timeout        int64
	timeoutReached bool
	response       common.IElements
	log            common.ILogger
}

func NewRequests() *Requests {
	this := &Requests{}
	this.pending = make(map[string]*Request)
	this.mtx = &sync.Mutex{}
	return this
}

func (this *Requests) NewRequest(msgNum uint32, msgSource string, timeout int64, log common.ILogger) *Request {
	request := &Request{}
	request.msgNum = msgNum
	request.msgSource = msgSource
	request.timeout = timeout
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
	delete(this.pending, key)
	return request
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

func (this *Request) Response() common.IElements {
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
	this.log.Info("Added timeout for request")
	time.Sleep(time.Second * time.Duration(this.timeout))
	this.log.Info("Checking timeout for request")
	this.Lock()
	defer this.Unlock()
	this.log.Info("After timeout for request")
	if this.response == nil {
		this.log.Info("Timeout reached for request")
		this.timeoutReached = true
		this.cond.Broadcast()
	}
}

func (this *Request) SetResponse(resp common.IElements) {
	this.response = resp
	this.cond.Broadcast()
}

func requestKey(msgSource string, msgNum uint32) string {
	key := bytes.Buffer{}
	key.WriteString(msgSource)
	key.WriteString(strconv.Itoa(int(msgNum)))
	return key.String()
}
