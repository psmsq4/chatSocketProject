package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== main 종료 시 자식 goroutine 강제 종료 실험 ===")

	go func() {
		fmt.Println("자식: 시작 - 10초 동안 실행 예정")
		for i := 0; i < 10; i++ {
			fmt.Printf("자식: %d\n", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println("자식: 완료") // 이것은 출력되지 않을 것
	}()

	fmt.Println("main: 3초 후 종료 예정")
	time.Sleep(3 * time.Second)
	fmt.Println("main: 종료 - 자식 goroutine도 강제 종료됨")
}






