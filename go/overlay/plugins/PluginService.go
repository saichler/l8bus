package plugins

import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8web"
)

const (
	ServiceName     = "Plugin"
	ServiceTypeName = "PluginService"
)

type PluginService struct {
}

func (this *PluginService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	vnic.Resources().Registry().Register(&l8web.L8Plugin{})
	return nil
}

func (this *PluginService) DeActivate() error {
	return nil
}

func (this *PluginService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	plugin := pb.Element().(*l8web.L8Plugin)
	err := LoadPlugin(plugin, vnic)
	if err != nil {
		vnic.Resources().Logger().Error(err.Error())
	}
	return object.New(err, nil)
}
func (this *PluginService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *PluginService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *PluginService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *PluginService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *PluginService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return object.New(nil, nil)
}
func (this *PluginService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

func (this *PluginService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

func (this *PluginService) WebService() ifs.IWebService {
	return nil
}
