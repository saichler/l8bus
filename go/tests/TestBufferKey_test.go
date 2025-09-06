package tests

import (
	"testing"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/layer8/go/overlay/vnic"
)

func TestBufferKey(t *testing.T) {
	key := vnic.BatchKey("Hello", 0, ifs.M_Leader)
	if key != "Hello04" {
		t.Fail()
	}
}
