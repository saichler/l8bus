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

package vnet

import (
	"github.com/saichler/l8types/go/ifs"
)

func (this *VNet) vnetServiceRequest(data []byte, vnic ifs.IVNic) {
	msg, err := this.protocol.MessageOf(data, this.resources)
	if err != nil {
		this.resources.Logger().Error(err)
		return
	}

	pb, err := this.protocol.ElementsOf(msg)
	if err != nil {
		if msg.Tr_State() != ifs.NotATransaction {
			//This message should not be processed and we should just
			//reply with nil to unblock the transaction
			vnic.Reply(msg, nil)
			return
		}
		this.resources.Logger().Error(err)
		return
	}

	// Otherwise call the handler per the action & the type
	if msg.Action() == ifs.Notify {
		resp := this.resources.Services().Notify(pb, vnic, msg, false)
		if resp != nil && resp.Error() != nil {
			panic(pb)
			this.resources.Logger().Error(resp.Error())
		}
		return
	}

	if msg.Reply() {
		this.vnic.SetResponse(msg, pb)
		return
	}
	var resp ifs.IElements
	if internal(msg) {
		resp = this.resources.Services().Handle(pb, msg.Action(), msg, this.vnic)
	} else {
		resp = this.resources.Services().Handle(pb, msg.Action(), msg, vnic)
	}
	if resp != nil && resp.Error() != nil {
		this.resources.Logger().Error(resp.Error(), " : ", msg.Action())
	}
	if msg.Request() {
		err = vnic.Reply(msg, resp)
		if err != nil {
			this.resources.Logger().Error(err.Error())
		}
	}
}

func (this *VNet) ExternalCount() int32 {
	return this.switchTable.conns.sizeExternalVnet.Load()
}

func (this *VNet) LocalCount() int32 {
	return this.switchTable.conns.sizeInternal.Load()
}

func internal(msg *ifs.Message) bool {
	if msg.Action() >= ifs.MapR_POST && msg.Action() <= ifs.MapR_GET {
		return true
	}
	if msg.ServiceName() == "users" || msg.ServiceName() == "roles" || msg.ServiceName() == "tokens" {
		return true
	}
	return false
}
