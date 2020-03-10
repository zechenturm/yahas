package service

import "testing"

func TestCreation(t *testing.T) {
	d := DependencyManager{}
	d.Add("test")
	if d.Order()[0] != "test" {
		t.Fatal("wrong order!")
	}
}
