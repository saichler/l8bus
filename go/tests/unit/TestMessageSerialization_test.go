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

package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
	"github.com/saichler/l8utils/go/utils/strings"
)

func testMessageSerialization(t *testing.T) {
	res, _ := t_resources.CreateResources(25000, 5, ifs.Trace_Level)
	size := 1000000
	start := time.Now().Unix()
	p := protocol.New(res)
	uuid := ifs.NewUuid()
	for i := 0; i < size; i++ {
		pb := &testtypes.TestProto{}
		pb.MyString = strings.New("Str-", i).String()
		pb.MyInt32 = int32(i)
		obj := object.New(nil, pb)
		d, _ := p.CreateMessageFor(uuid, "HelloWorld", 1, ifs.P1, ifs.M_All, ifs.POST, uuid, uuid, obj, false, false, 120,
			ifs.NotATransaction, "", "", -1, -1, -1, -1, -1, 0, false, "")
		msg, _ := p.MessageOf(d, res)
		p.ElementsOf(msg)
	}
	end := time.Now().Unix()
	fmt.Println((end - start))
}
