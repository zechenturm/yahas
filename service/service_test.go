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

func (service *testService) Additional() int {
	return 10
}

func TestSingleService(t *testing.T) {
	s := serviceManager{}

	serv := &testService{}

	s.Register("test", serv)

	serv2 := s.Get("test")

	if serv2 != serv {
		t.Fatal("service manager returned wrong service!")
	}

	if serv2.(*testService).Additional() != serv.Additional() {
		t.Fatal("Additional() returned wrong number!")
	}
}
