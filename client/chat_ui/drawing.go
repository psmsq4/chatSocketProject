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
var chatLogmutex sync.Mutex
var chatTransfermutex sync.Mutex
var chatLogCond sync.Cond
var msgInputCond sync.Cond
var _backup_message string
var _is_chat_transfer_signal bool

func DrawFrame(chatRoomID int16, messageListener chan string, chatLogBuffer []string) {
	var msg string
	chatLogCond.L = &chatLogmutex
	msgInputCond.L = &chatLogmutex

	_backup_message = ""
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
		_is_chat_transfer_signal = false
		msgInputCond.Broadcast()
		chatLogCond.L.Unlock()

		msg = <-messageListener // 이 코드는 DrawWriteMsgArea가 병행수행 중에도 실행될 수 있음.
		_is_chat_transfer_signal = true

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
	prefix := "New Message > "
	prefix_length := len(prefix)

	for {
		msgInputCond.L.Lock()
		for isChatLogCond {
			// lock을 잡고 isChatLogCond가 true인 경우 대기
			// 그런데 chatLogArea에서 채널에 블록되어 있으면 데드락 발생할 수도..
			// msg 입력 후 엔터 -> DrawWriteMsgArea에서 조건변수 true로 바꾸고 lock을 놓음.
			// -> 채팅 기록을 업데이트 함 -> ChatLogArea에서 조건변수 false로 바꾸고 lock을 놓고 채널에서 블록됨.
			// -> 다시 DrawWriteMsgArea가 lock을 잡음.
			// -> 채팅 기록을 업데이트 했음에도 불구하고 _is_chat_transfer_signal이 true임.
			// -> 그래서 1초 후 break 되고 _is_chat_transfer_signal을 false, 조건변수를 true로 바꾸로 lock을 놓음.
			// -> 그러나 chatLogArea는 채널에 블록되어 있는 상태이고 다시 DrawWriteMsgArea가 lock을 잡지만
			// -> 이전에 조건변수를 true로 바꿔놓는 바람에 Waiting이 강제되어 데드락이 발생함.

			// 위 문제를 해결하기 위해 chatLogArea에서 채팅 기록을 업데이트할 때 _is_chat_transfer_signal을 false로 바꿔놓으면?
			// 어차피 다음 채팅이 업데이트 되었으면 _is_chat_transfer_signal은 책임을 다한 것임.
			// -> DrawWriteMsgArea에서 1초 후 _is_chat_transfer_signal이 false이므로 break 될 일이 없음.
			// -> break 되지 않으니 엔터 또는 위/아래 화살표를 누르기 전까지는 메세지를 입력받게 됨.
			// -> 엔터를 눌러 break되면 서버에 전송되고 조건변수가 바뀌고 lock을 놓음.
			// -> 서버로부터 메세지가 전송되고 즉시 chatLogArea에서 매세지 채널로부터 수신받아 채팅 로그를 업데이트함
			// -> 채팅 로그 업데이트 시, _is_chat_transfer_signal을 false로 바꿔놓음. 그리고 채널에서 블록됨.
			// -> DrawWriteMsgArea는 무사히 Lock을 잡고 Waiting 없이 본인 업무를 수행함.
			// 이는 내가 전송한 메세지에 대해 발생하는 문제점에 대해 기술한 것임.
			// 만약 위 시나리오가 아니라 엔터를 눌러 서버에 전송하지 않고 다른 유저가 메세지를 보냈을 때도 유효할지 생각해봐야 함.
			// DrawWriteMsgArea에서 내가 GetChar()에 블록된 상태이고 1초마다 TimeOut된다고 가정.
			// -> 그 와중에 다른 유저가 메세지를 보내 채널에 도착함. 이때 병행수행 중인 ChatLogArea에서 채널 메세지를 수신하고 _is_chat_transfer_signal을 true로 변경함.
			// -> 나(DrawWriteMsgArea)는 1초 후 _is_chat_transfer_signal이 true임을 발견함.
			// -> break되고 조건변수를 true로 바꾸고 lock을 놓음.
			// -> lock을 놓고 바로 context switch가 발생할 수도 있고 아닐 수도 있지만 그건 일단 배제하고 chatLogArea에서 채팅 로그를 업데이트하고 _is_chat_transfer_signal을 false로 바꿈.
			// -> 그리고 chatLogArea에서 조건변수를 false로 바꾸고 lock을 놓게 되면, 다시 DrawWriteMsgArea가 lock을 잡고 Waiting 없이 수행됨.
			// -> 이때도 _is_chat_transfer_signal이 false니 문제가 발생하지 않음.
			msgInputCond.Wait()
		}
		writeMsgArea.Erase()
		writeMsgArea.Border(gc.ACS_BULLET, gc.ACS_BULLET, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE)

		message := ""
		if _backup_message != "" {
			message = _backup_message
			_backup_message = ""
		} else {
			message = prefix
		}

		writeMsgArea.Timeout(1000)
		for {
			writeMsgArea.MovePrint(1, 1, message)
			writeMsgArea.Refresh()

			char := writeMsgArea.GetChar()
			if char == 0 {
				if _is_chat_transfer_signal {
					// 입력 시간 1초 초과 시, 채팅이 서버로부터 전송되었는지 확인함
					// 전송 확인 방법 _is_chat_transfer_signal을 확인하면 됨.
					// chatLogArea에서 messageListener 채널에 메세지가 들어왔는지를 확인하고
					// _is_chat_transfer_signal을 true로 변경할 것임.
					_backup_message = message // chatLogArea에 채팅 기록을 업데이트한 후 다시 복구가 필요하므로 백업함.
					break
				} else {
					continue
				}

			} else if char == gc.KEY_RETURN {
				break
			} else if char == gc.KEY_UP {
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

		if _is_chat_transfer_signal {
			isChatLogCond = true
			chatLogCond.Broadcast()
			msgInputCond.L.Unlock()
			continue
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
