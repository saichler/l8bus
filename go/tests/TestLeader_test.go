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

	"github.com/saichler/l8types/go/ifs"
)

func getLeader(uuid string) ifs.IVNic {
	all := topo.AllVnics()
	for _, nic := range all {
		if nic.Resources().SysConfig().LocalUuid == uuid {
			return nic
		}
	}
	panic("No Leader")
}

func TestLeader(t *testing.T) {
	/*
		eg2_3 := topo.VnicByVnetNum(2, 3)
		hc := health.Health(eg2_3.Resources())
		leaderBefore := hc.LeaderFor(ServiceName, 0)
		leader := getLeader(leaderBefore)
		leader.Shutdown()
		defer func() {
			topo.RenewVnic(leader.Resources().SysConfig().LocalAlias)
		}()
		time.Sleep(time.Second * 10)
		leaderAfter := hc.LeaderFor(ServiceName, 0)
		if leaderAfter == leaderBefore {
			Log.Fail(t, "Expected leader to change")
			return
		}*/
}
