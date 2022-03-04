package utils

import (
	"fmt"
	"testing"
)

type testClone struct {
	name string
	age  int
}

func TestClone(t *testing.T) {
	p := &testClone{name: "name", age: 19}
	s := *p
	s.age = 20
	fmt.Println(p.age)
}
