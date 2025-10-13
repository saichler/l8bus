package health

import (
	"github.com/saichler/l8services/go/services/generic"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/web"
)

const (
	ServiceTypeName = "HealthService"
	ServiceName     = "Health"
	ServiceArea     = byte(0)
)

func Activate(vnic ifs.IVNic) {
	serviceConfig := &ifs.ServiceConfig{}
	serviceConfig.ServiceName = ServiceName
	serviceConfig.ServiceArea = ServiceArea

	services := &l8services.L8Services{}
	services.ServiceToAreas = make(map[string]*l8services.L8ServiceAreas)
	services.ServiceToAreas[ServiceName] = &l8services.L8ServiceAreas{}
	services.ServiceToAreas[ServiceName].Areas = make(map[int32]bool)
	services.ServiceToAreas[ServiceName].Areas[int32(ServiceArea)] = true

	serviceConfig.ServiceItem = &l8health.L8Health{AUuid: vnic.Resources().SysConfig().LocalUuid, Services: services}
	serviceConfig.InitItems = []interface{}{serviceConfig.ServiceItem}

	serviceConfig.SendNotifications = true
	serviceConfig.Transaction = false
	serviceConfig.PrimaryKey = []string{"AUuid"}
	serviceConfig.WebServiceDef = web.New(ServiceName, ServiceArea,
		nil, nil,
		nil, nil,
		nil, nil,
		nil, nil,
		&l8web.L8Empty{}, &l8health.L8Top{})
	generic.Activate(serviceConfig, vnic)
}

func HealthOf(uuid string, r ifs.IResources) *l8health.L8Health {
	sh, ok := HealthService(r)
	if ok {
		filter := &l8health.L8Health{}
		filter.AUuid = uuid
		h := sh.Get(object.New(nil, filter), nil)
		result, _ := h.Element().(*l8health.L8Health)
		return result
	}
	return nil
}

func HealthService(r ifs.IResources) (ifs.IServiceHandler, bool) {
	return r.Services().ServiceHandler(ServiceName, ServiceArea)
}

func HealthServiceCache(r ifs.IResources) (ifs.IServiceHandlerCache, bool) {
	hs, _ := HealthService(r)
	hc, ok := hs.(ifs.IServiceHandlerCache)
	return hc, ok
}

func All(r ifs.IResources) map[string]*l8health.L8Health {
	hc, _ := HealthServiceCache(r)
	col := hc.Collect(all)
	result := make(map[string]*l8health.L8Health)
	for _, h := range col {
		hp := h.(*l8health.L8Health)
		result[hp.AUuid] = hp
	}
	return result
}

func all(i interface{}) (bool, interface{}) {
	return true, i
}
