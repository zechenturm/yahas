package service

import "github.com/zechenturm/yahas/yahasplugin"

type serviceManager struct {
	service yahasplugin.Service
}

func (sm *serviceManager) Register(name string, service yahasplugin.Service) {
	sm.service = service
}

func (sm *serviceManager) Get(name string) yahasplugin.Service {
	return sm.service
}
