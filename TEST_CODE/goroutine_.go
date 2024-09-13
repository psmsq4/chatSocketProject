package main

import (
	"fmt"
	"time"
)

func say(s string)	{
	for i := 0; i < 10; i++	{
		fmt.Println(s, "***", i)
	}
}

func main()	{
	say("Sync")

	go say("Async1")
	go say("Async2")
	go say("Async3")

	time.Sleep(time.Second * 3)
}
