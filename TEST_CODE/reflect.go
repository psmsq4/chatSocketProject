package main

import	(
	"fmt"
	"reflect"
)

func main()	{
	var x = 42
	p := &x
	v := reflect.ValueOf(p)

	fmt.Println("Type of p:", v.Type())
	fmt.Println("Kind of p:", v.Kind())

	e := v.Elem()
	fmt.Println("Type of e:", e.Type())
	fmt.Println("Kind of e:", e.Kind())
	fmt.Println("Value of e:", e.Int())

	var i interface{} = 3.14
	v = reflect.ValueOf(i)
	fmt.Println("Type of i:", v.Type())
	fmt.Println("Kind of i:", v.Kind())

	e = v.Elem()
	fmt.Println("Type of e:", e.Type())
	fmt.Println("Kind of e:", e.Kind())
	fmt.Println("Value of e:", e.Float())
}
