package plugins

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"os"
	"plugin"
	"sync"
)

var loadedPlugins = make(map[string]*plugin.Plugin)
var mtx = &sync.Mutex{}

func loadPluginFile(p *types.Plugin) (*plugin.Plugin, error) {

	md5 := md5.New()
	md5Hash := base64.StdEncoding.EncodeToString(md5.Sum([]byte(p.Data)))
	mtx.Lock()
	defer mtx.Unlock()
	pluginFile, ok := loadedPlugins[md5Hash]
	if ok {
		return pluginFile, nil
	}

	data, err := base64.StdEncoding.DecodeString(p.Data)
	if err != nil {
		return nil, err
	}
	name := ifs.NewUuid() + ".so"
	err = os.WriteFile(name, data, 0777)
	if err != nil {
		return nil, err
	}
	defer os.Remove(name)

	pluginFile, err = plugin.Open(name)
	if err != nil {
		return nil, errors.New("failed to load plugin #1 " + err.Error())
	}

	loadedPlugins[md5Hash] = pluginFile

	return pluginFile, nil
}

func loadPlugin(p *types.Plugin, vnic ifs.IVNic, iRegistry, iService bool) error {
	pluginFile, err := loadPluginFile(p)
	if err != nil {
		return err
	}

	plg, err := pluginFile.Lookup("Plugin")
	if err != nil {
		return errors.New("failed to load plugin #2")
	}
	if plg == nil {
		return errors.New("failed to load plugin #3")
	}
	pluginInterface := *plg.(*ifs.IServicePlugin)
	if iRegistry {
		err = pluginInterface.InstallRegistry(vnic)
		if err != nil {
			return err
		}
	}
	if iService {
		err = pluginInterface.InstallServices(vnic)
	}
	return err
}
