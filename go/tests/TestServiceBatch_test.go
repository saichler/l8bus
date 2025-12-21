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

	"github.com/saichler/l8test/go/infra/t_service"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
)

func TestServiceBatch(t *testing.T) {
	vnic := topo.VnicByVnetNum(1, 1)
	link := ifs.NewServiceLink("", t_service.ServiceName, 0, 0, ifs.M_Proximity, 2, false)
	vnic.RegisterServiceLink(link)
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	vnic.Proximity(t_service.ServiceName, 0, ifs.PATCH, &testtypes.TestProto{MyString: "Hello"})
	time.Sleep(time.Second * 3)
	count := 0
	for _, v := range topo.AllHandlers() {
		count += v.PatchN()
	}
	if count != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1")
	}
}
