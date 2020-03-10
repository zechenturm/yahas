package service

type DependencyManager struct {
	deps []string
}

func (dm *DependencyManager) Add(name string) {
	dm.deps = append(dm.deps, name)
}

func (dm *DependencyManager) Order() []string {
	return dm.deps
}
