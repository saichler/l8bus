package protocol

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
)

var MessageLog bool = false
var MsgLog = newMessageTypeLog()
var started bool = false

type MessageTypeLog struct {
	mtx  sync.Mutex
	msgs map[string]int
}

func newMessageTypeLog() *MessageTypeLog {
	return &MessageTypeLog{msgs: make(map[string]int), mtx: sync.Mutex{}}
}

func (this *MessageTypeLog) AddLog(serviceName string, serviceArea byte, action ifs.Action) {
	if !MessageLog {
		return
	}
	key := strings.New(serviceName, serviceArea, action).String()
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if !started {
		started = true
		go this.log()
	}
	this.msgs[key]++
}

func (this *MessageTypeLog) Print() {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	for k, v := range this.msgs {
		fmt.Println(k, " - ", v)
	}
}

func (this *MessageTypeLog) log() {
	for {
		os.WriteFile("/tmp/log.csv", this.CSV(), 0777)
		time.Sleep(time.Second)
	}
}

func (this *MessageTypeLog) CSV() []byte {
	str := strings.New()
	str.Add("\"Key\",\"Count\"\n")
	this.mtx.Lock()
	defer this.mtx.Unlock()
	for k, v := range this.msgs {
		str.Add("\"")
		str.Add(k)
		str.Add("\",")
		str.Add(strconv.Itoa(v))
		str.Add("\n")
	}
	return str.Bytes()
}
