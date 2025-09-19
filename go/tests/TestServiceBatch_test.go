package tests

import (
	"testing"
	"time"

	"github.com/saichler/l8test/go/infra/t_service"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
)

func TestServiceBatch(t *testing.T) {
	vnic := topo.VnicByVnetNum(1, 1)
	link := ifs.NewServiceLink("", t_service.ServiceName, 0, 0, ifs.M_Proximity, 2, false)
	vnic.RegisterServiceLink(link)
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	time.Sleep(time.Second * 3)
	count := 0
	for _, v := range topo.AllHandlers() {
		count += v.PatchN()
	}
	if count != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1")
	}
}
