package accesscontrol

type Operation struct {
	ID         string `yaml:"id"`
	RBACAction string `yaml:"action"`
}

type Module struct {
	Name       string      `yaml:"name"`
	Operations []Operation `yaml:"operations"`
}

type RBACFile struct {
	Modules []Module `yaml:"modules"`
}
