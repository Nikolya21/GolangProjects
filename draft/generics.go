package main

import "fmt"

func equal[T int | float64](a, b T) bool {
	return a == b
}

type MyType int

func equalWithAlias[T ~int | ~float64](a, b T) bool {
	return a == b
}

func main() {
	fmt.Println(equalWithAlias(MyType(1), MyType(1)))
}
