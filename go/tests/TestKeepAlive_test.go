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

	"github.com/saichler/l8bus/go/overlay/health"
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_service"
	"github.com/saichler/l8types/go/ifs"
)

func TestKeepAlive(t *testing.T) {
	allVnics := topo.AllVnics()
	for _, nic := range allVnics {
		nic.Resources().SysConfig().KeepAliveIntervalSeconds = 2
	}

	defer func() {
		for _, nic := range allVnics {
			nic.Resources().SysConfig().KeepAliveIntervalSeconds = 30
		}
	}()

	pb := CreateTestModelInstance(3)
	eg2_1 := topo.VnicByVnetNum(2, 1)
	eg1_2 := topo.VnicByVnetNum(1, 2)
	err := eg2_1.Multicast(ServiceName, 0, ifs.POST, pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	time.Sleep(time.Second * time.Duration(eg2_1.Resources().SysConfig().KeepAliveIntervalSeconds+5))
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			nic := topo.VnicByVnetNum(i, j)
			hp := health.HealthOf(nic.Resources().SysConfig().LocalUuid, nic.Resources())
			if hp.Stats == nil {
				nic.Resources().Logger().Fail(t, "no stats for ", nic.Resources().SysConfig().LocalAlias)
				return
			}
		}
	}
	hp := health.HealthOf(eg2_1.Resources().SysConfig().LocalUuid, eg1_2.Resources())
	if hp.Stats.TxMsgCount == 0 {
		Log.Fail(t, "Expected at least one message to be sent for ", eg2_1.Resources().SysConfig().LocalUuid)
	}
}
