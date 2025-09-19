package tests

import (
	"testing"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8bus/go/overlay/vnic"
)

func TestBufferKey(t *testing.T) {
	key := vnic.LinkKeyByAttr("Hello", 0, ifs.M_Leader, false)
	if key != "Hello04false" {
		t.Fail()
	}
}
