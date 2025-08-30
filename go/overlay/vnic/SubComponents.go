package vnic

type SubComponents struct {
	components map[string]SubComponent
}

type SubComponent interface {
	name() string
	start()
	shutdown()
}

func newSubomponents() *SubComponents {
	egComponents := &SubComponents{}
	egComponents.components = make(map[string]SubComponent)
	return egComponents
}

func (egComponents *SubComponents) start() {
	for _, component := range egComponents.components {
		component.start()
	}
}

func (egComponents *SubComponents) shutdown() {
	// Shutdown in specific order: TX first, then RX, then others
	if tx := egComponents.components["TX"]; tx != nil {
		tx.shutdown()
	}
	if rx := egComponents.components["RX"]; rx != nil {
		rx.shutdown()
	}
	// Then shutdown remaining components
	for name, component := range egComponents.components {
		if name != "TX" && name != "RX" {
			component.shutdown()
		}
	}
}

func (egComponents *SubComponents) addComponent(component SubComponent) {
	egComponents.components[component.name()] = component
}

func (egComponents *SubComponents) TX() *TX {
	return egComponents.components["TX"].(*TX)
}
