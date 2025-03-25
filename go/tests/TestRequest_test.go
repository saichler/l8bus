package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_servicepoints"
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
)

func TestRequest(t *testing.T) {
	defer reset("TestRequest")
	pb := &testtypes.TestProto{MyString: "request"}
	eg3_1 := topo.VnicByVnetNum(3, 1)
	eg1_2 := topo.VnicByVnetNum(1, 2)
	resp, err := eg3_1.Request(eg1_2.Resources().Config().LocalUuid, ServiceName, 0, types.Action_POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	if resp.(*testtypes.TestProto).MyString != "request" {
		Log.Fail(t, "Expected response to be 'request")
		return
	}

	handler := topo.HandlerByVnetNum(1, 2)

	if handler.PostN() != 1 {
		Log.Fail(t, "eg1_2", " Post count does not equal 1")
		return
	}
}
