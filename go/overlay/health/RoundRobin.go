package health

import "github.com/saichler/l8types/go/ifs"

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
