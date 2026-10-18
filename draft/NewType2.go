package main

import (
	"fmt"
	"sort"
)

type Student struct {
	Age  int
	Name string
}

type Students []Student

func (sts Students) Len() int           { return len(sts) }
func (sts Students) Swap(a, b int)      { sts[a], sts[b] = sts[b], sts[a] }
func (sts Students) Less(a, b int) bool { return sts[a].Age < sts[b].Age }

func main() {
	sts := Students{
		{Age: 10, Name: "Иван"},
		{Age: 15, Name: "Петр"},
		{Age: 12, Name: "Саша"},
	}
	sort.Sort(sts)
	for _, st := range sts {
		fmt.Println(st)
	}
}
