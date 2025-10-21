package health

import (
	"github.com/saichler/l8services/go/services/base"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8utils/go/utils/web"
)

const (
	ServiceName = "Health"
)

func Activate(vnic ifs.IVNic, voter bool, secondary ...bool) {
	serviceArea := byte(0)
	if secondary != nil && secondary[0] {
		serviceArea = 1
	}
	serviceConfig := ifs.NewServiceLevelAgreement(&base.BaseService{}, ServiceName, serviceArea, true, nil)

	services := &l8services.L8Services{}
	services.ServiceToAreas = make(map[string]*l8services.L8ServiceAreas)
	services.ServiceToAreas[ServiceName] = &l8services.L8ServiceAreas{}
	services.ServiceToAreas[ServiceName].Areas = make(map[int32]bool)
	services.ServiceToAreas[ServiceName].Areas[int32(serviceArea)] = true

	serviceConfig.SetServiceItem(&l8health.L8Health{AUuid: vnic.Resources().SysConfig().LocalUuid, Services: services})
	serviceConfig.SetServiceItemList(&l8health.L8HealthList{})
	serviceConfig.SetInitItems([]interface{}{serviceConfig.ServiceItem()})

	serviceConfig.SetVoter(voter)
	serviceConfig.SetTransactional(false)
	serviceConfig.SetPrimaryKeys("AUuid")
	serviceConfig.SetWebService(web.New(ServiceName, serviceArea,
		nil, nil,
		nil, nil,
		nil, nil,
		nil, nil,
		&l8api.L8Query{}, &l8health.L8HealthList{}))
	base.Activate(serviceConfig, vnic)
}

func HealthOf(uuid string, r ifs.IResources, secondary ...bool) *l8health.L8Health {
	sh, ok := HealthService(r, secondary...)
	if ok {
		filter := &l8health.L8Health{}
		filter.AUuid = uuid
		h := sh.Get(object.New(nil, filter), nil)
		result, _ := h.Element().(*l8health.L8Health)
		return result
	}
	return nil
}

func HealthService(r ifs.IResources, secondary ...bool) (ifs.IServiceHandler, bool) {
	if secondary != nil && secondary[0] {
		return r.Services().ServiceHandler(ServiceName, 1)
	}
	return r.Services().ServiceHandler(ServiceName, 0)
}

func HealthServiceCache(r ifs.IResources, secondary ...bool) (ifs.IServiceHandlerCache, bool) {
	hs, _ := HealthService(r, secondary...)
	hc, ok := hs.(ifs.IServiceHandlerCache)
	return hc, ok
}

func All(r ifs.IResources, secondary ...bool) map[string]*l8health.L8Health {
	hc, _ := HealthServiceCache(r, secondary...)
	all := hc.All()
	result := make(map[string]*l8health.L8Health)
	for _, h := range all {
		hp := h.(*l8health.L8Health)
		result[hp.AUuid] = hp
	}
	return result
}
