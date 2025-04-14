package protocol

import "sync/atomic"

var CountMessages = false
var messagesCreated atomic.Uint64
var propertyChangeCalled atomic.Uint64

func AddMessageCreated() {
	if CountMessages {
		messagesCreated.Add(1)
	}
}

func AddPropertyChangeCalled() {
	if CountMessages {
		propertyChangeCalled.Add(1)
	}
}

func MessagesCreated() uint64 {
	return messagesCreated.Load()
}

func PropertyChangedCalled() uint64 {
	return propertyChangeCalled.Load()
}
