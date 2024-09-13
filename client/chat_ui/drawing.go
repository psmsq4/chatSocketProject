package chatui

import (
	"client/send_request"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	gc "github.com/gbin/goncurses"
)

const (
	TNewline = 1
)

func DrawFrame(chatRoomID int16, messageListener chan string, chatLogFileDescriptor *os.File) {
	var msg string

	chatLog := make([]byte, 1024)

	// filename := fmt.Sprintf("chatlog_%d", chatRoomID)

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

	max_y, max_x := stdscr.MaxYX()
	height_chatLogArea := max_y / 10 * 7
	height_wirteMsgArea := max_y / 10 * 3

	startRow := 0
	endRow := height_chatLogArea - 1

	chatLogArea, _ := gc.NewWindow(height_chatLogArea, max_x-20, 0, 20)

	defer chatLogArea.Delete()
	writeMsgArea, _ := gc.NewWindow(height_wirteMsgArea+1, max_x, height_chatLogArea, 0)
	defer writeMsgArea.Delete()

	go DrawWriteMsgArea(chatRoomID, writeMsgArea, messageListener)

	for {
		chatLogArea.Erase()
		chatLogArea.MovePrint(1, 0, string(chatLog))
		chatLogArea.Border(gc.ACS_BULLET, gc.ACS_BULLET, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE)

		chatLogArea.Refresh()
		chatLogArea.Keypad(false)
		msg = <-messageListener

		// nextLine += 2
		chatLog = make([]byte, 1024*10)

		if strings.Compare(msg, "quit") == 0 {
			break
		} else if strings.Compare(msg, "/c up") == 0 {
			chatLogFileDescriptor.ReadAt(chatLog, 0)
			startRow--
			endRow--
			chatLog = []byte(ExtractValidVolumnChatlog(startRow, endRow, string(chatLog)))
		} else if strings.Compare(msg, "/c down") == 0 {
			chatLogFileDescriptor.ReadAt(chatLog, 0)
			startRow++
			endRow++
			chatLog = []byte(ExtractValidVolumnChatlog(startRow, endRow, string(chatLog)))
		} else {
			chatLogFileDescriptor.WriteString(msg)
			chatLogFileDescriptor.ReadAt(chatLog, 0)
			chatLog = []byte(ExtractValidVolumnChatlog(startRow, endRow, string(chatLog)))
		}
	}
}

func ExtractValidVolumnChatlog(startRow, endRow int, originChatlog string) string {
	targetStrings := make([]string, 0)

	startIndex := 0

	for currentIndex, char := range originChatlog {
		if char == '\n' {
			targetStrings = append(targetStrings, originChatlog[startIndex:currentIndex])
			startIndex = currentIndex + 1
		}
	}

	if startRow < 0 {
		startRow = 0
	}
	if endRow > len(targetStrings) {
		endRow = len(targetStrings)
	}

	convertedString := strings.Join(targetStrings[startRow:endRow], "\n")
	targetStrings = nil
	return convertedString
}

func DrawWriteMsgArea(chatRoomID int16, writeMsgArea *gc.Window, messageListener chan string) {
	time.Sleep(10 * time.Millisecond)
	// var buffer string

	for {
		writeMsgArea.Erase()
		writeMsgArea.Border(gc.ACS_BULLET, gc.ACS_BULLET, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_HLINE)
		writeMsgArea.MovePrint(1, 1, "New Message > ")
		writeMsgArea.Refresh()
		message, err := writeMsgArea.GetString(512)
		if err != nil {
			panic(err)
		}
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
	}
}
