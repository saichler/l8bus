package protocol

import (
	"github.com/saichler/shared/go/share/logger"
	"github.com/saichler/types/go/types"
	"sync/atomic"
)

var CountMessages = false
var messagesCreated atomic.Uint64
var propertyChangeCalled atomic.Uint64
var ExplicitLog = logger.NewLoggerDirectImpl(logger.NewFileLogMethod("/tmp/Explicit.log"))

func AddMessageCreated() {
	if CountMessages {
		messagesCreated.Add(1)
	}
}

func AddPropertyChangeCalled(set *types.NotificationSet) {
	if CountMessages {
		propertyChangeCalled.Add(1)
		ExplicitLog.Trace("*** Property Change: ", set.ServiceName, " ", set.Type.String(), ":")
	}
}

func MessagesCreated() uint64 {
	return messagesCreated.Load()
}

func PropertyChangedCalled() uint64 {
	return propertyChangeCalled.Load()
}
