package tests

import (
	"github.com/saichler/shared/go/tests"
	"github.com/saichler/shared/go/types"
	"testing"
	"time"
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
	_, err := eg3.Request(types.CastMode_Single, types.Action_POST, 0, "TestProto", pb)
	if err != nil {
		log.Fail(t, err.Error())
		return
	}
	time.Sleep(1 * time.Second)
	if tsps["eg2"].PostNumber != 1 {
		log.Fail(t, "Expected post to be 1 but it is ", tsps["eg2"].PostNumber)
	}
}
