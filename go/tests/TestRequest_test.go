package tests

import (
	. "github.com/saichler/shared/go/tests/infra"
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
)

func TestRequest(t *testing.T) {
	defer reset("TestRequest")
	pb := &testtypes.TestProto{MyString: "request"}
	resp, err := eg2.Request(eg3.Resources().Config().LocalUuid, ServiceName, 0, types.Action_POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	if resp.(*testtypes.TestProto).MyString != "request" {
		Log.Fail(t, "Expected response to be 'request")
		return
	}

	if tsps["eg3"].PostN() != 1 {
		Log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}
}
