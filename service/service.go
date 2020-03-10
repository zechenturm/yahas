package service

import "github.com/zechenturm/yahas/yahasplugin"

type ServiceManager struct {
	services map[string]yahasplugin.Service
}

func (sm *ServiceManager) Register(name string, service yahasplugin.Service) {
	if sm.services == nil {
		sm.services = make(map[string]yahasplugin.Service)
	}
	sm.services[name] = service
}

func (sm *ServiceManager) Get(name string) yahasplugin.Service {
	return sm.services[name]
}

func (sm *ServiceManager) Unregister(name string) {
	delete(sm.services, name)
}
