package main

import (
	"fmt"
)

type MySlice []int

func (sl MySlice) String() string {
	return fmt.Sprintf("Мой кастомный список: %v", []int(sl))
}

func (sl MySlice) Sed() int {
	return len(sl)
}

func main() {
	var sl MySlice = MySlice{1, 2, 3}
	fmt.Println(sl) // fmt под капотом проверяет объекты - реализуют ли они интерфейс Stringer - метод String - если да то выводит то что в String

}
