package requests

import (
	"bytes"
	"github.com/saichler/l8srlz/go/l8srlz/object"
	"github.com/saichler/l8types/go/ifs"
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
	response       ifs.IElements
	log            ifs.ILogger
}

func NewRequests() *Requests {
	this := &Requests{}
	this.pending = make(map[string]*Request)
	this.mtx = &sync.Mutex{}
	return this
}

func (this *Requests) NewRequest(msgNum uint32, msgSource string, timeout int64, log ifs.ILogger) *Request {
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
	time.Sleep(time.Second * time.Duration(this.timeout))
	this.Lock()
	defer this.Unlock()
	if this.response == nil {
		this.timeoutReached = true
		this.cond.Broadcast()
	}
}

func (this *Request) SetResponse(resp ifs.IElements) {
	this.response = resp
	this.cond.Broadcast()
}

func requestKey(msgSource string, msgNum uint32) string {
	key := bytes.Buffer{}
	key.WriteString(msgSource)
	key.WriteString(strconv.Itoa(int(msgNum)))
	return key.String()
}
