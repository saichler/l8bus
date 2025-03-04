package tests

import (
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/types"
	"testing"
)

func TestTransaction(t *testing.T) {
	defer reset("TestTransaction")
	for _, ts := range tsps {
		ts.Tr = true
	}
	defer func() {
		for _, ts := range tsps {
			ts.Tr = false
		}
	}()
	pb := &tests.TestProto{MyString: "test"}
	resp, err := eg3.Request(types.CastMode_Single, types.Action_POST, 0, "TestProto", pb)
	if err != nil {
		log.Fail(t, err.Error())
		return
	}
	tr := resp.(*types.Tr)
	if tr.State != types.TrState_Commited {
		log.Fail(t, "transaction state is not commited,", tr.State.String())
		return
	}

	resp, err = eg3.Request(types.CastMode_Single, types.Action_POST, 0, "TestProto", pb)
	if err != nil {
		log.Fail(t, err.Error())
		return
	}
	tr = resp.(*types.Tr)
	if tr.State != types.TrState_Commited {
		log.Fail(t, "transaction state is not commited,", tr.State.String(), ":", tr.Id)
		return
	}

	if tsps["eg1"].PostNumber != 2 {
		log.Fail(t, "Expected post to be 2 but it is ", tsps["eg1"].PostNumber)
	}
	if tsps["eg2"].PostNumber != 2 {
		log.Fail(t, "Expected post to be 2 but it is ", tsps["eg2"].PostNumber)
	}
	if tsps["eg3"].PostNumber != 2 {
		log.Fail(t, "Expected post to be 2 but it is ", tsps["eg3"].PostNumber)
	}
	if tsps["eg4"].PostNumber != 2 {
		log.Fail(t, "Expected post to be 2 but it is ", tsps["eg4"].PostNumber)
	}
}

func sendTransaction(nic interfaces.IVirtualNetworkInterface, t *testing.T) {
	pb := &tests.TestProto{MyString: "test"}
	_, err := nic.Request(types.CastMode_Single, types.Action_POST, 0, "TestProto", pb)
	if err != nil {
		log.Fail(t, err.Error())
		return
	}
}
