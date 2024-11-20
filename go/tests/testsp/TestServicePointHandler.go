package testsp

import (
	"github.com/saichler/shared/go/share/interfaces"
	"github.com/saichler/shared/go/tests"
	"google.golang.org/protobuf/proto"
)

type TestServicePointHandler struct {
	PostNumber int
}

const (
	TEST_TOPIC = "Tests"
)

func NewTestServicePointHandler(registry interfaces.IStructRegistry, sp interfaces.IServicePoints) *TestServicePointHandler {
	tsp := &TestServicePointHandler{}
	registry.RegisterStruct(&tests.TestProto{})
	sp.RegisterServicePoint(&tests.TestProto{}, tsp, registry)
	return tsp
}

func (tsp *TestServicePointHandler) Post(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	interfaces.Logger().Debug("Test POST from ", edge.Config().Local_Uuid)
	tsp.PostNumber++
	return nil, nil
}
func (tsp *TestServicePointHandler) Put(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (tsp *TestServicePointHandler) Patch(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (tsp *TestServicePointHandler) Delete(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (tsp *TestServicePointHandler) Get(pb proto.Message, edge interfaces.IEdge) (proto.Message, error) {
	return nil, nil
}
func (tsp *TestServicePointHandler) EndPoint() string {
	return "/EdgeInfos"
}
func (tsp *TestServicePointHandler) Topic() string {
	return TEST_TOPIC
}
