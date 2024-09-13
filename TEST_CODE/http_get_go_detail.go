package main

import	(
	"fmt"
	"io/ioutil"
	"net/http"
)

func main()	{
	// Request 객체 생성
	req, err := http.NewRequest("GET", "http://naver.com", nil)
	if err != nil	{
		panic(err)
	}

	// 필요시 헤더 추가 가능
	req.Header.Add("User-Agent", "Crawler")

	// Client 객체에서 Request 실행
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil	{
		panic(err)
	}
	defer resp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes)
	fmt.Println(str)
}
