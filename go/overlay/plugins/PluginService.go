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
	Vnic ifs.IVNic
}

func (this *PluginService) Activate(serviceName string, serviceArea uint16,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	resources.Registry().Register(&types.Plugin{})
	return nil
}

func (this *PluginService) DeActivate() error {
	return nil
}

func (this *PluginService) Post(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	plugin := pb.Element().(*types.Plugin)
	err := loadPlugin(plugin, this.Vnic)
	if err != nil {
		resourcs.Logger().Error(err.Error())
	}
	return object.New(err, nil)
}
func (this *PluginService) Put(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *PluginService) Patch(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *PluginService) Delete(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *PluginService) GetCopy(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return nil
}
func (this *PluginService) Get(pb ifs.IElements, resourcs ifs.IResources) ifs.IElements {
	return object.New(nil, nil)
}
func (this *PluginService) Failed(pb ifs.IElements, resourcs ifs.IResources, msg ifs.IMessage) ifs.IElements {
	return nil
}

func (this *PluginService) TransactionMethod() ifs.ITransactionMethod {
	return nil
}
