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

package vnic

import "github.com/saichler/l8types/go/ifs"

func (this *VirtualNetworkInterface) Multicast(serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	return this.multicast(ifs.P8, ifs.M_All, serviceName, serviceArea, action, any)
}

func (this *VirtualNetworkInterface) multicast(priority ifs.Priority, multicastMode ifs.MulticastMode, serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	if this.serviceLinks != nil {
		key := LinkKeyByAttr(serviceName, serviceArea, multicastMode, false)
		exist, ok := this.serviceLinks.Load(key)
		if ok {
			link := exist.(*txServiceLink)
			if link.BatchMode() {
				link.Send(action, any)
				return nil
			}
		}
	}
	elems, err := createElements(any, this.resources)
	if err != nil {
		return err
	}
	return this.components.TX().Multicast("", serviceName, serviceArea, action, elems, priority, multicastMode,
		false, false, this.protocol.NextMessageNumber(), ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")
}

func (this *VirtualNetworkInterface) multicastLink(priority ifs.Priority, multicastMode ifs.MulticastMode, serviceName string, serviceArea byte, action ifs.Action, any interface{}) error {
	elems, err := createElements(any, this.resources)
	if err != nil {
		return err
	}
	return this.components.TX().Multicast("", serviceName, serviceArea, action, elems, priority, multicastMode,
		false, false, this.protocol.NextMessageNumber(), ifs.NotATransaction, "", "",
		-1, -1, -1, -1, -1, 0, false, "")
}
