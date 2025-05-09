package unit

import (
	"fmt"
	"github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/l8srlz/go/l8srlz/object"
	"github.com/saichler/l8utils/go/utils/strings"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
	"testing"
	"time"
)

func testMessageSerialization(t *testing.T) {
	res, _ := t_resources.CreateResources(25000, 5, ifs.Trace_Level)
	size := 1000000
	start := time.Now().Unix()
	p := protocol.New(res)
	uuid := ifs.NewUuid()
	for i := 0; i < size; i++ {
		pb := &testtypes.TestProto{}
		pb.MyString = strings.New("Str-", i).String()
		pb.MyInt32 = int32(i)
		obj := object.New(nil, pb)
		d, _ := p.CreateMessageFor(uuid, "HelloWorld", 1, ifs.P1, ifs.POST, uuid, uuid, obj, false, false, 120, nil)
		msg, _ := p.MessageOf(d)
		p.ElementsOf(msg)
	}
	end := time.Now().Unix()
	fmt.Println((end - start))
}
