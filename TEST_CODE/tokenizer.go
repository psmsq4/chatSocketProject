package main

import (
	"fmt"
	"strings"
)

const (
	TNewline = 1
)

func main()	{
	testStr := ` SSH  |  2024-08-23 15:42:14
 [0]: a

 SSH  |  2024-08-23 15:42:14
 [1]: b

 SSH  |  2024-08-23 15:42:14
 [2]: c

 SSH  |  2024-08-23 15:42:15
 [3]: d

 SSH  |  2024-08-23 15:42:15
 [4]: e

 SSH  |  2024-08-23 15:42:15
 [5]: f

 SSH  |  2024-08-23 15:42:16
 [6]: g

 SSH  |  2024-08-23 15:42:52
 [7]: /cup

`	
	targetStrings := make([]string, 0, 0)

	idx := 0
	startIdx := 0

	testByte := make([]byte, 10000)
	copy(testByte[:], testStr[:])

	for currentIdx, char := range string(testByte)	{
		if char == '\n'	{
			targetStrings = append(targetStrings, testStr[startIdx : currentIdx])
			startIdx = currentIdx + 1
			idx++
		}
	}

	for idx, str := range targetStrings	{
		fmt.Println("[", idx, "]: ", str)
	}

	convertedString := strings.Join(targetStrings[3:5], "\n")
	fmt.Println(convertedString)
	fmt.Println(targetStrings)
}
