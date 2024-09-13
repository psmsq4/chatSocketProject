package main

func main()	{
	ch := make(chan int, 2)

	// 채널에 송신
	ch <- 1
	ch <- 2

	// 채널을 닫는다. 닫아도 수신은 가능하다.
	close(ch)

	// 채널로부터 수신
	println(<-ch)
	println(<-ch)

	if _, success  := <-ch; !success	{
		println("더 이상 데이터 없음.")
	}
}
