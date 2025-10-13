package health

import (
	"github.com/saichler/l8services/go/services/generic"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
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
	serviceConfig.ServiceItem = &l8health.L8Health{}
	serviceConfig.SendNotifications = false
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
	sh, ok := r.Services().ServiceHandler(ServiceName, ServiceArea)
	if ok {
		filter := &l8health.L8Health{}
		filter.AUuid = uuid
		h := sh.Get(object.New(nil, filter), nil)
		result, _ := h.Element().(*l8health.L8Health)
		return result
	}
	return nil
}
