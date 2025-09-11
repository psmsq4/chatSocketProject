package chatui

import (
	"client/send_request"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	gc "github.com/gbin/goncurses"
)

const (
	TNewline = 1
)

var isChatLogCond = true // 채팅 로그 조건 변수
var mutex sync.Mutex
var chatLogCond sync.Cond
var msgInputCond sync.Cond

func DrawFrame(chatRoomID int16, messageListener chan string, chatLogBuffer []string) {
	var msg string
	chatLogCond.L = &mutex
	msgInputCond.L = &mutex

	chatLog := make([]byte, 1024)

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	gc.CBreak(true)
	gc.Cursor(0)

	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer gc.End()

	if !gc.HasColors() {
		log.Fatal("This requires a colour capable terminal")
	}

	if err := gc.StartColor(); err != nil {
		log.Fatal(err)
	}

	stdscr.ScrollOk(true)

	max_y, max_x := stdscr.MaxYX()
	height_chatLogArea := max_y / 10 * 7
	height_writeMsgArea := max_y / 10 * 3

	startRow := 0

	chatLogArea, _ := gc.NewWindow(height_chatLogArea, max_x-20, 0, 20)
	defer chatLogArea.Delete()
	writeMsgArea, _ := gc.NewWindow(height_writeMsgArea+1, max_x, height_chatLogArea, 0)
	defer writeMsgArea.Delete()

	chatLogArea.ScrollOk(true)
	writeMsgArea.ScrollOk(true)

	go DrawWriteMsgArea(chatRoomID, writeMsgArea, messageListener)

	for {
		chatLogCond.L.Lock()
		for !isChatLogCond {
			chatLogCond.Wait()
		}
		chatLogArea.Erase()
		chatLogArea.MovePrint(1, 0, string(chatLog))
		chatLogArea.Border(gc.ACS_BULLET, gc.ACS_BULLET, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE)

		chatLogArea.Refresh()
		chatLogArea.Keypad(false)

		isChatLogCond = false
		msgInputCond.Broadcast()
		chatLogCond.L.Unlock()

		msg = <-messageListener

		if strings.Compare(msg, "quit") == 0 {
			break
		} else if strings.Compare(msg, "/c up") == 0 {
			startRow--
			chatLog = []byte(ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer))
		} else if strings.Compare(msg, "/c down") == 0 {
			startRow++
			chatLog = []byte(ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer))
		} else {
			chatLogBuffer = append(chatLogBuffer, msg)

			startRow = -1

			chatLog = []byte(ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer))
		}
	}
}

func ExtractValidVolumnChatlog(startRow *int, chatLogAreaHeight int, chatLogBuffer []string) string {
	/*
		chatLogBuffer의 메세지 하나는 라인 4개를 차지함.
	*/

	chatLogAreaHeight = (chatLogAreaHeight - 2) / 4 // 상하 테두리 제외 + 4로 나누어서 몇 개의 메세지를 출력할 수 있는지 계산.

	if *startRow < 0 { // /c up 시 시작 인덱스가 0보다 작아지는 경우 또는 새 메세지가 도착했을 때, 마지막 채팅 기록으로 이동
		if len(chatLogBuffer)-chatLogAreaHeight < 0 {
			*startRow = 0
		} else {
			/* 전체 메세지 개수가 100개라 가정하자.
			 * 그리고 실제 채팅창 높이가 1이라 가정하자.
			 * 그러면 화면에 출력되어야 할 메세지는 [99, 100) == [99, 99] == 99번째 인덱스에 있는 메세지이다.
			 * 따라서 시작 인덱스를 99로 설정해야 한다.
			 * 이는 높이가 n이어도 성립한다.
			 */
			*startRow = len(chatLogBuffer) - chatLogAreaHeight
		}
	} else if *startRow > (len(chatLogBuffer)-1) && len(chatLogBuffer) > 1 {
		// /c down 시 시작 인덱스가 마지막 채팅 기록보다 큰 경우 마지막 채팅 기록으로 이동
		*startRow = len(chatLogBuffer) - 1
	}

	endRow := *startRow + chatLogAreaHeight
	if endRow > len(chatLogBuffer) {
		endRow = len(chatLogBuffer)
	}

	convertedString := strings.Join(chatLogBuffer[*startRow:endRow], "\n") // go도 python처럼 끝인덱스는 포함하지 않는다.
	chatLogBuffer = nil
	return convertedString
}

func DrawWriteMsgArea(chatRoomID int16, writeMsgArea *gc.Window, messageListener chan string) {
	time.Sleep(10 * time.Millisecond)
	// var buffer string

	writeMsgArea.Keypad(true)
	for {
		msgInputCond.L.Lock()
		for isChatLogCond {
			msgInputCond.Wait()
		}
		writeMsgArea.Erase()
		writeMsgArea.Border(gc.ACS_BULLET, gc.ACS_BULLET, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE)
		// message, err := writeMsgArea.GetString(512)
		prefix := "New Message > "
		message := prefix
		prefix_length := len(prefix)
		for {
			writeMsgArea.MovePrint(1, 1, message)
			writeMsgArea.Refresh()

			char := writeMsgArea.GetChar()
			if char == gc.KEY_RETURN {
				break
			}

			if char == gc.KEY_UP {
				message = "New Message > /c up"
				break
			} else if char == gc.KEY_DOWN {
				message = "New Message > /c down"
				break
			} else if char == gc.KEY_BACKSPACE {
				if len(message) > prefix_length {
					runes := []rune(message)
					runes = runes[:len(runes)-1]
					message = string(runes)
				}
			} else {
				message += string(char)
			}
		}

		message = message[prefix_length:]

		// if err != nil {
		// 	panic(err)
		// }
		if strings.Compare(message, "quit") == 0 {
			messageListener <- "quit"
			break
		} else if strings.Compare(message, "/c up") == 0 {
			messageListener <- "/c up"
		} else if strings.Compare(message, "/c down") == 0 {
			messageListener <- "/c down"
		} else {
			// 입력받은 메세지를 서버로 전송해야함.
			send_request.SendTransferMessage(chatRoomID, message)
		}
		isChatLogCond = true
		chatLogCond.Broadcast()
		msgInputCond.L.Unlock()
	}
}
