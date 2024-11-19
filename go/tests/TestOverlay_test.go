package tests

import (
	"github.com/saichler/layer8/go/overlay/edge"
	"github.com/saichler/shared/go/share/interfaces"
	"testing"
	"time"
)

func TestOverlay(t *testing.T) {
	defer shutdownTopology()
	egImpl := eg1.(*edge.EdgeImpl)
	time.Sleep(time.Second * 10)
	interfaces.Info("*****************************************************************")
	egImpl.State()
	interfaces.Info("*****************************************************************")
}
