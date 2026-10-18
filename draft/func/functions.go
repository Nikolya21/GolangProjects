package main

import (
	"fmt"
)

func main() {
	func(in string) {
		fmt.Println("printer outs:", in)
	}("hello")
	printer := func(in string) {
		fmt.Println("printer outs:", in)
	}
	printer("hello")

	type strFuncType func(string)
	worker := func(callback strFuncType) {
		callback("hello")
	}

	worker(printer)

	func(in ...any) {
		fmt.Println(in)
	}(42, "cjev")
	var a sa
	a = "21"
	a = 10
	fmt.Println(a)
}

type sa interface{}
