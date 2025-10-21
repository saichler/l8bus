package tests

import (
	"testing"

	vnet2 "github.com/saichler/l8bus/go/overlay/vnet"
	infra "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
)

func TestExternalVnic(t *testing.T) {
	r, _ := infra.CreateResources(53555, 0, ifs.Debug_Level)
	vnet := vnet2.NewVNet(r, true)
	vnet.Start()
}
