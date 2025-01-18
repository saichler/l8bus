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
	for _, component := range egComponents.components {
		component.shutdown()
	}
}

func (egComponents *SubComponents) addComponent(component SubComponent) {
	egComponents.components[component.name()] = component
}

func (egComponents *SubComponents) TX() *TX {
	return egComponents.components["TX"].(*TX)
}
