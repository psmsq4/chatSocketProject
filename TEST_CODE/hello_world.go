package main

import "fmt"

func test(a int, b int) (c int, d int) {
	c = a
	d = b
	return
}

func funcParm(method func(a int) int, a int) int {
	for i := range a {
		fmt.Println(i)
	}
	return method(a)
}

func main() {
	var a, b int
	c := 10

	fmt.Println("hello World!")

	a, b = test(10, 20)

	fmt.Println(a, b)
	fmt.Println(c)

	var anony = func(a int) int {
		return a
	}

	fmt.Println(funcParm(anony, 100))
}
