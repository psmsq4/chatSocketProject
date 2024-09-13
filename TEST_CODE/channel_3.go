package main

import "fmt"

func main()	{
	ch := make(chan string, 1);
	sendChan(ch)
	receiveChan(ch)
}

func sendChan(ch chan <- string)	{
	ch <- "Data"
	// x := <- ch // 에러발생 :: 송신 채널에 대해서 수신을 시도했기 때문에
}

func receiveChan(ch <- chan string)	{
	data := <- ch
	fmt.Println(data)
}
