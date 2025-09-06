package plugins

import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
)

const (
	ServiceName     = "Plugin"
	ServiceTypeName = "PluginService"
)

type PluginService struct {
}

func (this *PluginService) Activate(serviceName string, serviceArea byte,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	resources.Registry().Register(&types.Plugin{})
	return nil
}

func (this *PluginService) DeActivate() error {
	return nil
}

func (this *PluginService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	plugin := pb.Element().(*types.Plugin)
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
