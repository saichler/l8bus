package protocol

import (
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"sync/atomic"
)

var CountMessages = false
var messagesCreated atomic.Uint64
var propertyChangeCalled atomic.Uint64

func AddMessageCreated() {
	if CountMessages {
		messagesCreated.Add(1)
	}
}

func AddPropertyChangeCalled(vnic common.IVirtualNetworkInterface, set *types.NotificationSet) {
	if CountMessages {
		propertyChangeCalled.Add(1)
		vnic.Resources().Logger().Trace("*** Property Change: ", set.ServiceArea, " ", set.Type.String(), ":")
	}
}

func MessagesCreated() uint64 {
	return messagesCreated.Load()
}

func PropertyChangedCalled() uint64 {
	return propertyChangeCalled.Load()
}
