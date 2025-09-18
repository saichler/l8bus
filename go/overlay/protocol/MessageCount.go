package protocol

import (
	"bytes"
	"github.com/saichler/l8utils/go/utils/logger"
	"sync/atomic"
)

var CountMessages = false
var messagesCreated atomic.Uint64
var handleData atomic.Uint64
var propertyChangeCalled atomic.Uint64
var ExplicitLog = logger.NewLoggerDirectImpl(logger.NewFileLogMethod("/tmp/Explicit.log"))

func AddMessageCreated() {
	if CountMessages {
		messagesCreated.Add(1)
	}
}

func AddPropertyChangeCalled(set *types.NotificationSet, alias string) {
	if CountMessages {
		propertyChangeCalled.Add(1)
		props := ""
		if set.Type == types.NotificationType_Update {
			buff := bytes.Buffer{}
			buff.WriteString(" - ")
			for _, chg := range set.NotificationList {
				buff.WriteString(chg.PropertyId)
				buff.WriteString(" ")
			}
			props = buff.String()
		}
		ExplicitLog.Trace("*** Property Change: ", alias, " ", set.ServiceName, " ", set.Type.String(), ":", props)
	}
}

func MessagesCreated() uint64 {
	return messagesCreated.Load()
}

func PropertyChangedCalled() uint64 {
	return propertyChangeCalled.Load()
}

func AddHandleData() {
	if CountMessages {
		handleData.Add(1)
	}
}

func HandleData() uint64 {
	return handleData.Load()
}
