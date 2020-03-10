package service

type DependencyManager struct {
}

func (*DependencyManager) Add(name string) {

}

func (*DependencyManager) Order() []string {
	return []string{"test"}
}
