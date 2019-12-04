package main

import "fmt"

type Test struct {
	str string
}

func (test *Test) String() string {
	res := fmt.Sprintf("hello world:%s\n", test.str)
	return res
}

func main() {
	t1 := &Test{"nihao"}
	fmt.Printf("%v\n", t1)
}
