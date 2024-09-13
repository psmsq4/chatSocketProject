package main

import (
	"fmt"
	"encoding/json"
)

func main()	{
	slcD := []string{"apple", "peach", "pear"}
	slcB, _ := json.Marshal(slcD)
	fmt.Println(string(slcB))

	str_slice := make([]string, 3)
	json.Unmarshal(slcB, &str_slice)
	fmt.Println(str_slice)
}
