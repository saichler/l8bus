package vnic

type KeepAlive struct {
	vnic *VirtualNetworkInterface
}

func (this *KeepAlive) start()    {}
func (this *KeepAlive) shutdown() {}
func (this *KeepAlive) name() string {
	return "KA"
}
func (this *KeepAlive) run() {
	for this.vnic.running {

	}
}
