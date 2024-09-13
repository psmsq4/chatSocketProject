package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Person
type Person struct {
	Name string
	Age  int
}

func main() {
	person := Person{"Alex", 10}
	pbytes, _ := json.Marshal(person)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post("http://httpbin.org/post", "application/json", buff)

	resp_body, err := io.ReadAll(resp.Body)
	if err == nil {
		println(string(resp_body))
	}

	defer resp.Body.Close()
}
