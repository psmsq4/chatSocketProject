package main

import (
	"os"
	"os/exec"
)

func main() {
	//정수형 채널을 생성한다.
	ch := make(chan int)

	go func() {
		ch <- 123
	}()

	var i int
	i = <-ch
	println(i)

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
