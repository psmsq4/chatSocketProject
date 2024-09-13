package main

import	(
	"fmt"
	"sync"
)

func main()	{
	// WaitGroup 생성. 2개의 Go루틴을 기다림.
	var wait sync.WaitGroup
	wait.Add(2)

	// 익명함수를 사용한 goroutine
	go func()	{
		defer wait.Done() // 끝나면 .Done() 호출
		fmt.Println("Hello")
	} () // <- 정의하자마자 바로 사용

	// 익명함수에 파라미터 전달
	go func(msg string)	{
		defer wait.Done()
		fmt.Println(msg)
	} ("Hi")

	wait.Wait() // Go루틴이 모두 끝날 때까지 대기
}
