package service

import (
	"testing"
)

type testService struct {
	Num int
}

func (service *testService) Name() string {
	return "test"
}
func (service *testService) ProvidedBy() string {
	return "test"
}

func (service *testService) Additional() int {
	return service.Num
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

func TestMultiplrServices(t *testing.T) {
	s := serviceManager{}

	serv1 := &testService{Num: 10}
	serv2 := &testService{Num: 20}

	s.Register("test1", serv1)
	s.Register("test2", serv2)

	if s.Get("test1") != serv1 {
		t.Fatal("Get() returned wrong test!")
	}

	if serv1.Additional() != s.Get("test1").(*testService).Additional() {
		t.Fatal("Additional() returned wrong number!")
	}

	if serv2.Additional() != s.Get("test2").(*testService).Additional() {
		t.Fatal("Additional() returned wrong number!")
	}

}
