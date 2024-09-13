package main

import (
	"encoding/json"
	"fmt"
)

type Member struct {
	Name   string
	Age    int
	Active bool
}

func main() {
	// go 데이터
	mem := Member{"Alex", 10, true}

	jsonBytes, err := json.Marshal(mem)
	if err != nil {
		panic(err)
	}

	// JSON 바이트를 문자열로 변경
	jsonString := string(jsonBytes)
	fmt.Println(jsonString)
}
