package main

import (
	"fmt"
)

type Dog interface {
	Greet()
}

var _ Dog = (*DogA)(nil)

type DogA struct {
	Name string
}

func (d *DogA) Greet() {
	fmt.Println("greet", d.Name)
}

var _ Dog = (*DogB)(nil)

type DogB struct {
	DogA
}

func (d *DogB) Greet() {
	fmt.Println("DogB greet", d.Name)
}

func main() {
	dogs := []Dog{&DogA{Name: "ken"}, &DogB{DogA{Name: "bow"}}}
	for _, d := range dogs {
		d.Greet()
	}
}
