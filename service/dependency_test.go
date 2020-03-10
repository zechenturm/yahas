package service

import "testing"

func TestCreation(t *testing.T) {
	d := DependencyManager{}
	d.Add("test")
	if d.Order()[0] != "test" {
		t.Fatal("wrong order!")
	}
}

func TestSimple(t *testing.T) {
	d := DependencyManager{}
	d.Add("test1")
	d.Add("test2")

	order := d.Order()

	if len(order) != 2 {
		t.Fatal("wrong number of arguments")
	}

	if d.Order()[0] != "test1" {
		t.Fatal("wrong 1st element!")
	}

	if d.Order()[1] != "test2" {
		t.Fatal("wrong 2nd element!")
	}

}
