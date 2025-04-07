package unit

import (
	"fmt"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/shared/go/share/strings"
	"github.com/saichler/types/go/testtypes"
	"testing"
	"time"
)

func testMessageSerialization(t *testing.T) {
	//res, _ := t_resources.CreateResources(25000, 5)
	size := 1000000
	start := time.Now().Unix()
	//p := protocol.New(res)
	//uuid := common.NewUuid()
	for i := 0; i < size; i++ {
		pb := &testtypes.TestProto{}
		pb.MyString = strings.New("Str-", i).String()
		pb.MyInt32 = int32(i)
		obj := object.New(nil, pb)
		obj.Serialize()
		//p.CreateMessageFor(uuid, "HelloWorld", 1, common.P1, common.POST, uuid, uuid, obj, false, false, 120, nil)
	}
	end := time.Now().Unix()
	fmt.Println((end - start))
}
