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

package tests

import (
	"testing"
	"time"

	vnet2 "github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8bus/go/overlay/vnic"
	infra "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
)

func TestExternalVnic(t *testing.T) {
	r, _ := infra.CreateResources(53555, 0, ifs.Debug_Level)
	vnet := vnet2.NewVNet(r, true)
	vnet.Start()

	r, _ = infra.CreateResources(53555, 1, ifs.Debug_Level)
	nic1 := vnic.NewVirtualNetworkInterface(r, nil)

	r, _ = infra.CreateResources(53555, 2, ifs.Debug_Level)
	r.SysConfig().RemoteVnet = "127.0.0.1"
	nic2 := vnic.NewVirtualNetworkInterface(r, nil)

	nic1.Start()
	nic1.WaitForConnection()
	nic2.Start()
	nic2.WaitForConnection()
	time.Sleep(time.Second)
}
