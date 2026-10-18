package main

import (
	"cmp"
	"fmt"
)

var (
	a = 1
	b = 2
)

type Number interface {
	~int | ~float64
}

type Box[T any] struct {
	Value T
}

func (b *Box[T]) String() string {
	return fmt.Sprintf("Box<%T>", b.Value)
}

type Calculator[T Number] struct {
	Value T
}

func (c Calculator[T]) Add(a T) T {
	return c.Value + a
}

func Compare[T cmp.Ordered](a, b T) bool {
	return a == b
}

func main() {
	item := &Box[int]{}
	fmt.Println(item)
	fmt.Println(Compare(1, 2))
}
