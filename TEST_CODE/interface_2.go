package main

import (
	"fmt"
	"reflect"
)

func main() {
	var x interface{}
	x = 1
	x = "Tom"

	printIt(x)
}

func printIt(v interface{}) {
	println(v.(string))
	fmt.Println(v)

	type_x := reflect.TypeOf(v)
	fmt.Println(type_x)

	kind_x := type_x.Kind()
	fmt.Println(kind_x)
}
