//go:build scale

package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_servicepoints"
	"github.com/saichler/shared/go/share/strings"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/testtypes"
	"testing"
	"time"
)

/*
func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}*/

func scaleTest(size, exp int, timeout int64, t *testing.T) bool {
	start := time.Now().Unix()
	eg2 := topo.VnicByVnetNum(1, 2)
	eg3 := topo.VnicByVnetNum(2, 1)
	for i := 0; i < size; i++ {
		pb := &testtypes.TestProto{}
		pb.MyString = strings.New("Str-", i).String()
		pb.MyInt32 = int32(i)
		err := eg2.Unicast(eg3.Resources().SysConfig().LocalUuid, ServiceName, 0, common.POST, pb)
		if err != nil {
			Log.Fail(t, err)
			return false
		}
	}

	handler := topo.HandlerByVnetNum(2, 1)

	now := time.Now().Unix()
	for handler.PostN() < exp {
		if time.Now().Unix()-timeout >= now {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	end := time.Now().Unix()
	Log.Info("Scale test for ", size, " took ", (end - start), " seconds")
	if handler.PostN() != exp {
		Log.Fail(t, "eg3", " Post count does not equal to ", exp, ":", handler.PostN())
		return false
	}
	return true
}

func TestScale(t *testing.T) {
	Log.SetLogLevel(common.Info_Level)
	exp := 1000
	ok := scaleTest(1000, exp, 4, t)
	if !ok {
		return
	}
	exp += 10000
	ok = scaleTest(10000, exp, 4, t)
	if !ok {
		return
	}
	exp += 100000
	ok = scaleTest(100000, exp, 10, t)
	if !ok {
		return
	}
	exp += 1000000
	scaleTest(1000000, exp, 20, t)
}
