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

	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_service"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
)

func TestRequest(t *testing.T) {
	defer reset("TestRequest")
	pb := &testtypes.TestProto{MyString: "request"}
	eg3_1 := topo.VnicByVnetNum(3, 1)
	eg1_2 := topo.VnicByVnetNum(1, 2)
	resp := eg3_1.Request(eg1_2.Resources().SysConfig().LocalUuid, ServiceName, 0, ifs.POST, pb, 5)
	if resp.Error() != nil {
		Log.Fail(t, resp.Error())
		return
	}

	if resp.Element().(*testtypes.TestProto).MyString != "request" {
		Log.Fail(t, "Expected response to be 'request")
		return
	}

	handler := topo.HandlerByVnetNum(1, 2)

	if handler.PostN() != 1 {
		Log.Fail(t, "eg1_2", " Post count does not equal 1")
		return
	}
}
