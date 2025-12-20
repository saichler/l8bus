package health

import (
	"fmt"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
)

type RoundRobin struct {
	participants []string
	index        int
}

func NewRoundRobin(serviceName string, serviceArea byte, r ifs.IResources) *RoundRobin {
	rr := &RoundRobin{}
	rr.participants = make([]string, 0)
	pMap := Participants(serviceName, serviceArea, r)
	for uuid, _ := range pMap {
		rr.participants = append(rr.participants, uuid)
	}
	return rr
}

func (this *RoundRobin) Next() string {
	if this.index >= len(this.participants) {
		this.index = 0
	}
	next := this.participants[this.index]
	this.index++
	return next
}

func Participants(serviceName string, serviceArea byte, r ifs.IResources) map[string]bool {
	hc, _ := HealthServiceCache(r)
	all := hc.All()
	result := make(map[string]bool)
	for _, h := range all {
		hp := h.(*l8health.L8Health)
		if hp.Services != nil && hp.Services.ServiceToAreas != nil {
			fmt.Println("[Round Robin] ServiceName ", serviceName)
			areas, ok := hp.Services.ServiceToAreas[serviceName]
			if ok && areas.Areas != nil {
				v, ok2 := areas.Areas[int32(serviceArea)]
				fmt.Println("[Round Robin] ServiceArea ", serviceArea, " ", v, " ", ok2)
				if ok2 {
					result[hp.AUuid] = true
				}
			}
		}
	}
	fmt.Println("[Round Robin] UUIDs ", len(result))
	return result
}
