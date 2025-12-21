// © 2025 Sharon Aicler (saichler@gmail.com)
//
// Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugins

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"os"
	"plugin"
	"sync"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/strings"
)

var loadedPlugins = make(map[string]*plugin.Plugin)
var mtx = &sync.Mutex{}

func loadPluginFile(p *l8web.L8Plugin) (*plugin.Plugin, error) {

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
	name := strings.New(ifs.NewUuid(), ".so").String()
	err = os.WriteFile(name, data, 0777)
	if err != nil {
		return nil, err
	}
	defer os.Remove(name)

	pluginFile, err = plugin.Open(name)
	if err != nil {
		return nil, errors.New(strings.New("failed to load plugin #1 ", err.Error()).String())
	}

	loadedPlugins[md5Hash] = pluginFile

	return pluginFile, nil
}

func LoadPlugin(p *l8web.L8Plugin, vnic ifs.IVNic) error {
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
	pluginInterface := *plg.(*ifs.IPlugin)
	err = pluginInterface.Install(vnic)
	if err != nil {
		return err
	}
	return err
}
