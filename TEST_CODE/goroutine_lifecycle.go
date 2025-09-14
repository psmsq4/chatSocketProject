package main

import (
	"fmt"
	"time"
)

func parentGoroutine() {
	fmt.Println("부모 goroutine 시작")

	// 자식 goroutine 생성
	go func() {
		fmt.Println("자식 goroutine 시작")
		for i := 0; i < 10; i++ {
			fmt.Printf("자식: %d\n", i)
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("자식 goroutine 완료")
	}()

	// 부모는 2초 후 종료
	time.Sleep(2 * time.Second)
	fmt.Println("부모 goroutine 종료")
}

func main() {
	fmt.Println("=== 실험 1: 부모 goroutine 종료 후 자식 상태 ===")

	go parentGoroutine()

	// main에서 충분히 기다려서 자식이 완료되는지 확인
	time.Sleep(8 * time.Second)

	fmt.Println("\n=== 실험 2: main 종료 시 자식 상태 ===")

	go func() {
		fmt.Println("새 자식 goroutine 시작")
		for i := 0; i < 20; i++ {
			fmt.Printf("새 자식: %d\n", i)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("새 자식 goroutine 완료")
	}()

	// main이 2초 후 종료되면 자식도 강제 종료
	time.Sleep(2 * time.Second)
	fmt.Println("main 종료 - 모든 goroutine 강제 종료")
}


