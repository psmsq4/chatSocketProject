package main

import "fmt"

func main()	{
	var a interface{} = 1

	i := a
	j := a.(int)

	fmt.Println(i)
	fmt.Println(j)
}
