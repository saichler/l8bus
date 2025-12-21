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

import "sync"

type RouteTable struct {
	routes   *sync.Map
	vnetUuid string
}

func newRouteTable(vnetUuid string) *RouteTable {
	return &RouteTable{routes: &sync.Map{}, vnetUuid: vnetUuid}
}

func (this *RouteTable) addRoutes(routes map[string]string) map[string]string {
	added := make(map[string]string)
	for k, v := range routes {
		_, ok := this.routes.Load(k)
		if !ok {
			this.routes.Store(k, v)
			added[k] = v
		}
	}
	return added
}

func (this *RouteTable) removeRoutes(toRemove map[string]string) map[string]string {
	removed := make(map[string]string)
	for uuid, _ := range toRemove {
		vnetUuid, ok := this.routes.Load(uuid)
		if ok {
			removed[uuid] = vnetUuid.(string)
		}
		this.routes.Range(func(k, v interface{}) bool {
			if uuid == v.(string) {
				removed[k.(string)] = v.(string)
			}
			return true
		})
	}
	for k, _ := range removed {
		this.routes.Delete(k)
	}
	return removed
}

func (this *RouteTable) vnetOf(uuid string) (string, bool) {
	vnetUuid, ok := this.routes.Load(uuid)
	if ok {
		return vnetUuid.(string), ok
	}
	return this.vnetUuid, false
}
