package plugins

import (
	"encoding/base64"
	"errors"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"os"
	"plugin"
)

func loadPlugin(p *types.Plugin, vnic ifs.IVNic) error {
	data, err := base64.StdEncoding.DecodeString(p.Data)
	if err != nil {
		return err
	}
	name := ifs.NewUuid() + ".so"
	err = os.WriteFile(name, data, 0777)
	if err != nil {
		return err
	}
	defer os.Remove(name)

	pluginFile, err := plugin.Open(name)
	if err != nil {
		return errors.New("failed to load plugin #1")
	}
	plugin, err := pluginFile.Lookup("Plugin")
	if err != nil {
		return errors.New("failed to load plugin #2")
	}
	if plugin == nil {
		return errors.New("failed to load plugin #3")
	}
	pluginInterface := *plugin.(*ifs.IPlugin)
	return pluginInterface.Install(vnic)
}
