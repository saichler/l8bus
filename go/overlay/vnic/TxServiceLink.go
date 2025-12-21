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

import (
	"sync"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8utils/go/utils/strings"
)

type txServiceLink struct {
	mtx      *sync.Mutex
	link     *l8services.L8ServiceLink
	queue    []*txServiceLinkEntry
	interval time.Duration
	vnic     *VirtualNetworkInterface
}

type txServiceLinkEntry struct {
	element interface{}
	action  ifs.Action
}

func (this *VirtualNetworkInterface) RegisterServiceLink(link *l8services.L8ServiceLink) {
	if this.serviceLinks == nil {
		this.serviceLinks = &sync.Map{}
	}
	key := LinkKeyByLink(link)
	_, ok := this.serviceLinks.Load(key)
	if !ok {
		this.serviceLinks.Store(key, newTxServiceLink(link, this))
	}
}

func newTxServiceLink(link *l8services.L8ServiceLink, vnic *VirtualNetworkInterface) *txServiceLink {
	tsb := &txServiceLink{}
	tsb.mtx = &sync.Mutex{}
	tsb.queue = make([]*txServiceLinkEntry, 0)
	tsb.link = link
	tsb.vnic = vnic
	tsb.interval = time.Duration(link.Interval)
	go tsb.watch()
	return tsb
}

func (this *txServiceLink) BatchMode() bool {
	return this.link.Interval > 0
}

func (this *txServiceLink) Send(action ifs.Action, element interface{}) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.queue = append(this.queue, &txServiceLinkEntry{element: element, action: action})
}

func (this *txServiceLink) watch() {
	for this.vnic.Running() {
		this.flush()
		time.Sleep(time.Second * this.interval)
	}
}

func (this *txServiceLink) flush() {
	this.mtx.Lock()
	items := this.queue
	this.queue = make([]*txServiceLinkEntry, 0)
	defer this.mtx.Unlock()
	if len(items) > 0 {
		var list []interface{}
		lastAction := -1
		for _, item := range items {
			if lastAction != int(item.action) {
				if list != nil {
					this.send(ifs.Action(lastAction), list)
				}
				list = make([]interface{}, 0)
				lastAction = int(item.action)
			}
			list = append(list, item.element)
		}
		if list != nil {
			this.send(ifs.Action(lastAction), list)
		}
	}
}

func (this *txServiceLink) send(action ifs.Action, elements []interface{}) {
	this.vnic.multicastLink(ifs.P7, ifs.MulticastMode(this.link.Mode),
		this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), action, elements)
}

func LinkKeyByLink(link *l8services.L8ServiceLink) string {
	return strings.New(link.ZsideServiceName, link.ZsideServiceArea, link.Mode, link.Request).String()
}

func LinkKeyByAttr(serviceName string, serviceArea byte, mode ifs.MulticastMode, request bool) string {
	return strings.New(serviceName, serviceArea, mode, request).String()
}
