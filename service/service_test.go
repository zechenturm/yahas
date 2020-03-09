package service

import (
	"testing"
)

type testService struct {
}

func (service *testService) Name() string {
	return "test"
}
func (service *testService) ProvidedBy() string {
	return "test"
}

func TestSingleService(t *testing.T) {
	s := serviceManager{}

	serv := &testService{}

	s.Register("test", serv)

	serv2 := s.Get("test")

	if serv2 != serv {
		t.Fatal("service manager returned wrong service!")
	}
}
