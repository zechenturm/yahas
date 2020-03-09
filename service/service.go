package service

import "github.com/zechenturm/yahas/yahasplugin"

type serviceManager struct {
	services map[string]yahasplugin.Service
}

func (sm *serviceManager) Register(name string, service yahasplugin.Service) {
	if sm.services == nil {
		sm.services = make(map[string]yahasplugin.Service)
	}
	sm.services[name] = service
}

func (sm *serviceManager) Get(name string) yahasplugin.Service {
	return sm.services[name]
}
