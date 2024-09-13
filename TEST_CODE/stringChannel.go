package main

import (
	"fmt"
	//"time"
)

var _MessageListener chan string

func main()	{
	_MessageListener = make(chan string)

	go routine1()

	_MessageListener <- "HI"
	//time.Sleep(2 * time.Second)
}

func routine1()	{
	msg := <- _MessageListener
	fmt.Println(msg)
}
