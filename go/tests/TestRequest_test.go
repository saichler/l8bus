package tests

import (
	"github.com/saichler/types/go/testtypes"
	"github.com/saichler/types/go/types"
	"testing"
)

func TestRequest(t *testing.T) {
	defer reset("TestRequest")
	pb := &testtypes.TestProto{MyString: "request"}
	resp, err := eg2.Request(types.CastMode_Single, types.Action_POST, 0, eg3.Resources().Config().LocalUuid, pb)
	if err != nil {
		log.Fail(t, err)
		return
	}

	if resp.(*testtypes.TestProto).MyString != "request" {
		log.Fail(t, "Expected response to be 'request")
		return
	}

	if tsps["eg3"].PostNumber != 1 {
		log.Fail(t, "eg3", " Post count does not equal 1")
		return
	}
}
