package chatui

import (
	"client/send_request"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	gc "github.com/gbin/goncurses"
)

const (
	TNewline = 1
)

// 전역변수로 상태를 관리하기에는 채팅방이 갱신될 때마다 오염 위험이 있음.
// 그러나 지역변수로 관리하기에는 함수 매개변수가 더러워짐.
// claude는 다음과 같이 상태 객체를 정의하고 채팅방이 생성될 때 상태 객체를 생성할 것을 권하고 있음.
// 내가 보기에도 괜찮은 방법임.
type ChatUIState struct {
	WindowMutex             sync.Mutex
	chatLogCond             sync.Cond
	msgInputCond            sync.Cond
	optionCond              sync.Cond
	wg                      *sync.WaitGroup
	isChatLogCond           bool
	isOptionCond            bool
	killSwitch              bool
	is_chat_transfer_signal bool
	backup_message          string
	chatLog                 string
}

// 아래 프로시저로 상태를 생성함.
// 이는 채팅방이 소멸할 때 같이 소멸하며 생성될 때 같이 생성됨.
func NewChatUIState() *ChatUIState {
	// Go에서 &ChatUIState{}는 ChatUIState 구조체의 인스턴스를 힙에 할당하고,
	// 그 주소값(포인터)을 반환한다.
	// 만약 state := ChatUIState{}처럼 값으로 선언하면 스택에 할당될 수 있지만,
	// &를 붙여 포인터로 생성하면 escape analysis에 의해 힙에 할당된다.
	// 반환 타입이 *ChatUIState이므로, 함수가 끝나도 state의 값이 복사되지 않고
	// 힙에 남아있는 동일한 객체의 포인터가 반환된다.
	// 즉, 반환 지점에서 값 복사가 일어나지 않는다.
	state := &ChatUIState{}

	state.chatLogCond.L = &state.WindowMutex
	state.msgInputCond.L = &state.WindowMutex
	state.optionCond.L = &state.WindowMutex
	state.wg = &sync.WaitGroup{}
	state.isChatLogCond = true
	state.isOptionCond = false
	state.killSwitch = false
	state.is_chat_transfer_signal = false
	state.backup_message = ""
	state.chatLog = ""

	return state
}

func DrawBeforeLogin(stdscr *gc.Window, UIListener chan string) {
	// build the menu items
	menu_items := []string{
		"Sign In",
		"Create Account"}
	items := make([]*gc.MenuItem, len(menu_items))
	for i, val := range menu_items {
		items[i], _ = gc.NewItem(val, "")
		defer items[i].Free()
	}

	// create the menu
	menu, _ := gc.NewMenu(items)
	defer menu.Free()

	MaxY, MaxX := stdscr.MaxYX()
	HEIGHT_OPTION_WIN := MaxY / 3
	WIDTH_OPTION_WIN := MaxX / 2

	menuwin, _ := gc.NewWindow(HEIGHT_OPTION_WIN, WIDTH_OPTION_WIN, MaxY/2-HEIGHT_OPTION_WIN/2, MaxX/2-WIDTH_OPTION_WIN/2)
	defer menuwin.Delete()
	menuwin.Keypad(true)
	menuwin.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)
	menu.SetWindow(menuwin)

	HEIGHT_OPTION_SUBWIN := HEIGHT_OPTION_WIN - 2
	WIDTH_OPTION_SUBWIN := WIDTH_OPTION_WIN - 2

	dwin := menuwin.Derived(HEIGHT_OPTION_SUBWIN-2, WIDTH_OPTION_SUBWIN, 3, 1)
	defer dwin.Delete()
	menu.SubWindow(dwin)
	menu.Format(5, 1)
	menu.Mark(" * ")

	title := "LifeGame Client v0.1"

	gc.InitPair(2, gc.C_RED, gc.C_BLACK)
	menuwin.AttrOn(gc.ColorPair(2) | gc.A_BOLD)
	menuwin.MovePrint(1, WIDTH_OPTION_SUBWIN/2-len(title)/2, title)
	menuwin.HLine(2, 1, 0, WIDTH_OPTION_WIN-2)
	menuwin.AttrOff(gc.ColorPair(2) | gc.A_BOLD)

	/* for init OptionArea Drawing */
	menu.Post()
	for {
		menuwin.Refresh()
		ch := menuwin.GetChar()
		menu.Driver(gc.DriverActions[ch])

		if ch == gc.KEY_RETURN {
			currentMenu := menu.Current(nil)
			fmt.Print(currentMenu.Name())
			if strings.Compare(currentMenu.Name(), "Sign In") == 0 {
				UIListener <- "Sign In"
				break
			} else if strings.Compare(currentMenu.Name(), "Create Account") == 0 {
				UIListener <- "Kill"
				break
			}
		}
	}
}

func DrawLogin(stdscr *gc.Window, UIListener chan string) (string, string) {
	MaxY, MaxX := stdscr.MaxYX()

	HEIGHT_LOGIN_WIN := MaxY / 3
	WIDTH_LOGIN_WIN := MaxX / 3

	Y_LOGIN_WIN := (MaxY - HEIGHT_LOGIN_WIN) / 2
	X_LOGIN_WIN := (MaxX - WIDTH_LOGIN_WIN) / 2

	gc.Echo(false)
	LoginWin, err := gc.NewWindow(HEIGHT_LOGIN_WIN, WIDTH_LOGIN_WIN, Y_LOGIN_WIN, X_LOGIN_WIN)
	defer LoginWin.Delete()
	if err != nil {
		log.Fatal("Fail to Create Login Window")
	}

	LoginWin.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)
	LoginWin.Keypad(true)
	stdscr.Keypad(true)

	HEIGHT_LOGIN_SUBWIN := HEIGHT_LOGIN_WIN - 2
	WIDTH_LOGIN_SUBWIN := WIDTH_LOGIN_WIN - 2
	// Y_LOGIN_SUBWIN := Y_LOGIN_WIN + 1
	// X_LOGIN_SUBWIN := X_LOGIN_WIN + 1

	fields := make([]*gc.Field, 2)
	fields[0], _ = gc.NewField(1, 16, HEIGHT_LOGIN_SUBWIN/2-1, 5+WIDTH_LOGIN_SUBWIN/2-10, 0, 0)
	defer fields[0].Free()
	fields[0].SetForeground(gc.ColorPair(1))
	fields[0].SetBackground(gc.A_UNDERLINE)
	fields[0].SetOptionsOff(gc.FO_AUTOSKIP)

	fields[1], _ = gc.NewField(1, 16, HEIGHT_LOGIN_SUBWIN/2+1, 5+WIDTH_LOGIN_SUBWIN/2-10, 0, 0)
	defer fields[1].Free()
	fields[1].SetForeground(gc.ColorPair(1))
	fields[1].SetBackground(gc.A_UNDERLINE)
	fields[1].SetOptionsOff(gc.FO_AUTOSKIP)

	form, _ := gc.NewForm(fields)
	form.SetWindow(LoginWin)
	// 중요!!!!!
	// : Derived된 자식 윈도우의 좌표값은 부모 윈도우의 상대 좌표이다.
	dwin := LoginWin.Derived(HEIGHT_LOGIN_SUBWIN, WIDTH_LOGIN_SUBWIN, 1, 1) // form.post 전에 form에 먼저 붙어야(set) 함.
	defer dwin.Delete()
	form.SetSub(dwin) // form.post 전에 form에 먼저 붙여야(set) 함.
	form.Post()
	defer form.UnPost()
	defer form.Free()
	dwin.Refresh()
	LoginWin.Refresh()

	gc.InitPair(1, gc.C_CYAN, gc.C_BLACK)
	/* fields는 dwin을 기준으로 생성되므로 text도 dwin에 그린다. */
	dwin.AttrOn(gc.ColorPair(1) | gc.A_BOLD)
	dwin.MovePrint(HEIGHT_LOGIN_SUBWIN/2-1, WIDTH_LOGIN_SUBWIN/2-10, "ID : ")
	dwin.MovePrint(HEIGHT_LOGIN_SUBWIN/2+1, WIDTH_LOGIN_SUBWIN/2-10, "PW : ")
	dwin.AttrOff(gc.ColorPair(1) | gc.A_BOLD)
	dwin.Refresh()
	LoginWin.Refresh()

	var id string
	var pw string
	is_id := true

	for {
		ch := LoginWin.GetChar()

		is_enter := false
		switch ch {
		case gc.KEY_BACKSPACE:
			if len(id)-1 < 0 {
				continue
			}
			form.Driver(gc.REQ_DEL_PREV)
			if is_id {
				id = id[0 : len(id)-1]
			} else {
				pw = pw[0 : len(pw)-1]
			}
		case gc.KEY_DOWN:
			form.Driver(gc.REQ_NEXT_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_id = !is_id
		case gc.KEY_UP:
			form.Driver(gc.REQ_PREV_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_id = !is_id
		case gc.KEY_TAB:
			form.Driver(gc.REQ_PREV_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_id = !is_id
		case gc.KEY_RETURN:
			form.Driver(gc.REQ_VALIDATION)
			is_enter = true
		default:
			// 인쇄 가능한 문자만 허용
			if ch >= 32 && ch <= 126 {
				form.Driver(ch)
				if is_id {
					id += string(ch)
				} else {
					pw += string(ch)
				}
			}
		}

		if is_enter {
			break
		}
	}

	stdscr.Clear()
	return id, pw
}

func DrawAfterLogin(stdscr *gc.Window, UIListener chan string, userID string) {
	// build the menu items
	menu_items := []string{
		"Find Chatroom",
		"View Involved In Chatroom",
		"Generate New Chat"}
	items := make([]*gc.MenuItem, len(menu_items))
	for i, val := range menu_items {
		items[i], _ = gc.NewItem(val, "")
		defer items[i].Free()
	}

	// create the menu
	menu, _ := gc.NewMenu(items)
	defer menu.Free()

	MaxY, MaxX := stdscr.MaxYX()
	HEIGHT_OPTION_WIN := MaxY / 3
	WIDTH_OPTION_WIN := MaxX / 2

	menuwin, _ := gc.NewWindow(HEIGHT_OPTION_WIN, WIDTH_OPTION_WIN, MaxY/2-HEIGHT_OPTION_WIN/2, MaxX/2-WIDTH_OPTION_WIN/2)
	defer menuwin.Delete()
	menuwin.Keypad(true)
	menuwin.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)
	menu.SetWindow(menuwin)

	HEIGHT_OPTION_SUBWIN := HEIGHT_OPTION_WIN - 2
	WIDTH_OPTION_SUBWIN := WIDTH_OPTION_WIN - 2

	dwin := menuwin.Derived(HEIGHT_OPTION_SUBWIN-2, WIDTH_OPTION_SUBWIN, 3, 1)
	defer dwin.Delete()
	menu.SubWindow(dwin)
	menu.Format(5, 1)
	menu.Mark(" * ")

	title := "Welcome " + userID + "!!"

	gc.InitPair(2, gc.C_RED, gc.C_BLACK)
	menuwin.AttrOn(gc.ColorPair(2) | gc.A_BOLD)
	menuwin.MovePrint(1, WIDTH_OPTION_SUBWIN/2-len(title)/2, title)
	menuwin.HLine(2, 1, 0, WIDTH_OPTION_WIN-2)
	menuwin.AttrOff(gc.ColorPair(2) | gc.A_BOLD)

	/* for init OptionArea Drawing */
	menu.Post()
	for {
		menuwin.Refresh()
		ch := menuwin.GetChar()
		menu.Driver(gc.DriverActions[ch])

		if ch == gc.KEY_RETURN {
			currentMenu := menu.Current(nil)
			fmt.Print(currentMenu.Name())
			if strings.Compare(currentMenu.Name(), "Find Chatroom") == 0 {
				UIListener <- "ChatList"
				break
			} else if strings.Compare(currentMenu.Name(), "View Involved In Chatroom") == 0 {
				UIListener <- "OldChat"
				break
			} else if strings.Compare(currentMenu.Name(), "Generate New Chat") == 0 {
				UIListener <- "NewChat"
				break
			}
		}
	}
	stdscr.Clear()
}

func DrawNewChat(stdscr *gc.Window, UIListener chan string) (string, string) {
	MaxY, MaxX := stdscr.MaxYX()

	HEIGHT_NEWCHAT_WIN := MaxY / 3
	WIDTH_NEWCHAT_WIN := MaxX / 3

	Y_NEWCHAT_WIN := (MaxY - HEIGHT_NEWCHAT_WIN) / 2
	X_NEWCHAT_WIN := (MaxX - WIDTH_NEWCHAT_WIN) / 2

	gc.Echo(false)
	NewChatWin, err := gc.NewWindow(HEIGHT_NEWCHAT_WIN, WIDTH_NEWCHAT_WIN, Y_NEWCHAT_WIN, X_NEWCHAT_WIN)
	defer NewChatWin.Delete()
	if err != nil {
		log.Fatal("Fail to Create NewChat Window")
	}

	NewChatWin.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)
	NewChatWin.Keypad(true)
	stdscr.Keypad(true)

	HEIGHT_NEWCHAT_SUBWIN := HEIGHT_NEWCHAT_WIN - 2
	WIDTH_NEWCHAT_SUBWIN := WIDTH_NEWCHAT_WIN - 2
	// Y_LOGIN_SUBWIN := Y_LOGIN_WIN + 1
	// X_LOGIN_SUBWIN := X_LOGIN_WIN + 1

	maxInput := 20
	field_chatname_label := "CHATROOM TITLE : "
	field_chatpw_label := "CHATROOM    PW : "

	fields := make([]*gc.Field, 2)
	fields[0], _ = gc.NewField(1, maxInput, HEIGHT_NEWCHAT_SUBWIN/2-1, WIDTH_NEWCHAT_SUBWIN/2-(len(field_chatname_label)+maxInput)/2+len(field_chatname_label), 0, 0)
	defer fields[0].Free()
	fields[0].SetForeground(gc.ColorPair(1))
	fields[0].SetBackground(gc.A_UNDERLINE)
	fields[0].SetOptionsOff(gc.FO_AUTOSKIP)

	fields[1], _ = gc.NewField(1, maxInput, HEIGHT_NEWCHAT_SUBWIN/2+1, WIDTH_NEWCHAT_SUBWIN/2-(len(field_chatpw_label)+maxInput)/2+len(field_chatpw_label), 0, 0)
	defer fields[1].Free()
	fields[1].SetForeground(gc.ColorPair(1))
	fields[1].SetBackground(gc.A_UNDERLINE)
	fields[1].SetOptionsOff(gc.FO_AUTOSKIP)

	form, _ := gc.NewForm(fields)
	form.SetWindow(NewChatWin)
	// 중요!!!!!
	// : Derived된 자식 윈도우의 좌표값은 부모 윈도우의 상대 좌표이다.
	dwin := NewChatWin.Derived(HEIGHT_NEWCHAT_SUBWIN, WIDTH_NEWCHAT_SUBWIN, 1, 1) // form.post 전에 form에 먼저 붙어야(set) 함.
	defer dwin.Delete()
	form.SetSub(dwin) // form.post 전에 form에 먼저 붙여야(set) 함.
	form.Post()
	defer form.UnPost()
	defer form.Free()
	dwin.Refresh()
	NewChatWin.Refresh()

	gc.InitPair(1, gc.C_CYAN, gc.C_BLACK)
	/* fields는 dwin을 기준으로 생성되므로 text도 dwin에 그린다. */
	dwin.AttrOn(gc.ColorPair(1) | gc.A_BOLD)
	dwin.MovePrint(HEIGHT_NEWCHAT_SUBWIN/2-1, WIDTH_NEWCHAT_SUBWIN/2-(len(field_chatname_label)+20)/2, field_chatname_label)
	dwin.MovePrint(HEIGHT_NEWCHAT_SUBWIN/2+1, WIDTH_NEWCHAT_SUBWIN/2-(len(field_chatpw_label)+20)/2, field_chatpw_label)
	dwin.AttrOff(gc.ColorPair(1) | gc.A_BOLD)
	dwin.Refresh()
	NewChatWin.Refresh()

	var chatname string
	var chatpw string
	is_chatname := true

	for {
		ch := NewChatWin.GetChar()

		is_enter := false
		switch ch {
		case gc.KEY_BACKSPACE:
			if len(chatname)-1 < 0 {
				continue
			}
			form.Driver(gc.REQ_DEL_PREV)
			if is_chatname {
				chatname = chatname[0 : len(chatname)-1]
			} else {
				chatpw = chatpw[0 : len(chatpw)-1]
			}
		case gc.KEY_DOWN:
			form.Driver(gc.REQ_NEXT_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_chatname = !is_chatname
		case gc.KEY_UP:
			form.Driver(gc.REQ_PREV_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_chatname = !is_chatname
		case gc.KEY_TAB:
			form.Driver(gc.REQ_PREV_FIELD)
			form.Driver(gc.REQ_END_LINE)
			is_chatname = !is_chatname
		case gc.KEY_RETURN:
			form.Driver(gc.REQ_VALIDATION)
			is_enter = true
		default:
			// 인쇄 가능한 문자만 허용
			if ch >= 32 && ch <= 126 {
				form.Driver(ch)
				if is_chatname {
					chatname += string(ch)
				} else {
					chatpw += string(ch)
				}
			}
		}

		if is_enter {
			break
		}
	}

	stdscr.Clear()
	return chatname, chatpw
}

func DrawChatRoom(stdscr *gc.Window, chatRoomID int16, chatLogBuffer []string, messageListener chan string, UIListener chan string) {
	chatUIState := NewChatUIState()

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	max_y, max_x := stdscr.MaxYX()
	height_chatLogArea := max_y / 10 * 7
	height_writeMsgArea := max_y / 10 * 3

	chatLogArea, _ := gc.NewWindow(height_chatLogArea, max_x-20, 0, 20)
	defer chatLogArea.Delete()
	writeMsgArea, _ := gc.NewWindow(height_writeMsgArea+1, max_x, height_chatLogArea, 0)
	defer writeMsgArea.Delete()

	chatLogArea.ScrollOk(true)
	writeMsgArea.ScrollOk(true)
	writeMsgArea.Keypad(true)
	/* ---------------------------------------------------------- */
	// build the menu items
	menu_items := []string{
		"Return Chat",
		"Exit Chat"}
	items := make([]*gc.MenuItem, len(menu_items))
	for i, val := range menu_items {
		items[i], _ = gc.NewItem(val, "")
		defer items[i].Free()
	}

	// create the menu
	menu, _ := gc.NewMenu(items)
	defer menu.Free()

	HEIGHT_OPTION_AREA := height_chatLogArea
	WIDTH_OPTION_AREA := 20

	menuwin, _ := gc.NewWindow(HEIGHT_OPTION_AREA, WIDTH_OPTION_AREA, 0, 0)
	defer menuwin.Delete()
	menuwin.Keypad(true)
	menuwin.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)

	menu.SetWindow(menuwin)
	dwin := menuwin.Derived(HEIGHT_OPTION_AREA-2, WIDTH_OPTION_AREA-2, 1, 1)
	defer dwin.Delete()
	menu.SubWindow(dwin)
	menu.Format(5, 1)
	menu.Mark(" * ")

	/* for init OptionArea Drawing */
	menu.Post()
	menu.Window().Refresh()

	chatUIState.wg.Add(3)
	/* ---------------------------------------------------------- */
	go DrawChatRoomOptionArea(menu, messageListener, chatUIState)
	go DrawChatLogArea(chatLogArea, messageListener, chatLogBuffer, height_chatLogArea, chatUIState)
	go DrawChatRoomInputArea(chatRoomID, writeMsgArea, messageListener, chatUIState)
	chatUIState.wg.Wait() // 현재 프로시저가 먼저 종료되어 버리면 goncurses 객체들의 메모리 해제로 Segmentation Fault가 발생한다. 따라서 Waiting 로직이 필요함.

	UIListener <- "AfterLogin"
	stdscr.Clear()
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

func DrawChatLogArea(chatLogArea *gc.Window, messageListener chan string, chatLogBuffer []string, height_chatLogArea int, state *ChatUIState) {
	var msg string

	startRow := 0

	for {
		state.chatLogCond.L.Lock()
		for !state.isChatLogCond || state.isOptionCond {
			state.chatLogCond.Wait()
		}
		chatLogArea.Erase()
		chatLogArea.MovePrint(1, 0, state.chatLog)
		chatLogArea.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)

		chatLogArea.Refresh()
		chatLogArea.Keypad(false)

		state.isChatLogCond = false
		state.is_chat_transfer_signal = false

		state.msgInputCond.Broadcast()
		state.chatLogCond.L.Unlock()

		msg = <-messageListener // 이 코드는 DrawWriteMsgArea가 병행수행 중에도 실행될 수 있음.
		state.is_chat_transfer_signal = true

		if strings.Compare(msg, "quit") == 0 {
			// time.Sleep(time.Second)
			break
		} else if strings.Compare(msg, "/c up") == 0 {
			startRow--
			state.chatLog = ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer)
		} else if strings.Compare(msg, "/c down") == 0 {
			startRow++
			state.chatLog = ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer)
		} else {
			chatLogBuffer = append(chatLogBuffer, msg)

			startRow = -1

			state.chatLog = ExtractValidVolumnChatlog(&startRow, height_chatLogArea, chatLogBuffer)
		}
	}

	state.wg.Done()
}

func DrawChatRoomOptionArea(optionArea *gc.Menu, messageListener chan string, state *ChatUIState) {
	/* 옵션창은 WriteMsgArea에서 특정 커맨드 키를 눌렀을 때 제어권을 넘겨받는 특수한 영역임.
	 * WriteMsgArea에서 옵션 커맨드가 눌리면, 일단 WriteMsgArea는 OptionCond를 true로 변경하고 Lock을 놓음.
	 * Unlock전에 BroadCast해서 Watiing 중인 OptionArea를 깨움.
	 * 이때 ChatLogArea가 깨어날 걱정은 안 해도 됨. ChatLogCond는 여전히 false이기 때문이다.
	 *  */

	for {
		state.optionCond.L.Lock()
		for !state.isOptionCond || state.isChatLogCond {
			state.optionCond.Wait()
		}

		optionArea.Post()

		for {
			optionArea.Window().Refresh()
			ch := optionArea.Window().GetChar()
			optionArea.Driver(gc.DriverActions[ch])

			if ch == gc.KEY_RETURN {
				currentMenu := optionArea.Current(nil)
				if strings.Compare(currentMenu.Name(), "Return Chat") == 0 {
					break
				} else if strings.Compare(currentMenu.Name(), "Exit Chat") == 0 {
					messageListener <- "quit"
					state.killSwitch = true
					break
				}
			}
		}

		state.isOptionCond = false
		state.msgInputCond.Broadcast()
		state.optionCond.L.Unlock()

		if state.killSwitch {
			break
		}
	}

	state.wg.Done()
}

func DrawChatRoomInputArea(chatRoomID int16, writeMsgArea *gc.Window, messageListener chan string, state *ChatUIState) {
	// var buffer string
	prefix := " New Message > "
	prefix_length := len(prefix)

	is_option_activated := false

	for {
		state.msgInputCond.L.Lock()
		for state.isChatLogCond || state.isOptionCond { // ChatLog가 활성화 되어 있거나 OptionCond가 활성화 되어 있거나
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
			state.msgInputCond.Wait()
		}

		message := ""
		if state.backup_message != "" {
			message = state.backup_message
			state.backup_message = ""
		} else {
			message = prefix
		}

		writeMsgArea.Timeout(1000)
		for {
			writeMsgArea.Erase()
			writeMsgArea.Border(gc.ACS_VLINE, gc.ACS_VLINE, gc.ACS_HLINE, gc.ACS_HLINE, gc.ACS_LLCORNER, gc.ACS_URCORNER, gc.ACS_ULCORNER, gc.ACS_LRCORNER)
			writeMsgArea.MovePrint(1, 1, message)
			writeMsgArea.Refresh()
			if state.killSwitch {
				break
			}
			char := writeMsgArea.GetChar()

			if char == 0 {
				if state.is_chat_transfer_signal {
					// 입력 시간 1초 초과 시, 채팅이 서버로부터 전송되었는지 확인함
					// 전송 확인 방법 _is_chat_transfer_signal을 확인하면 됨.
					// chatLogArea에서 messageListener 채널에 메세지가 들어왔는지를 확인하고
					// _is_chat_transfer_signal을 true로 변경할 것임.
					state.backup_message = message // chatLogArea에 채팅 기록을 업데이트한 후 다시 복구가 필요하므로 백업함.
					break
				} else {
					continue
				}

			} else if char == gc.KEY_RETURN {
				break
			} else if char == gc.KEY_UP {
				message = prefix + "/c up"
				break
			} else if char == gc.KEY_DOWN {
				message = prefix + "/c down"
				break
			} else if char == gc.KEY_F1 {
				is_option_activated = true
				state.backup_message = message
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

		if state.killSwitch {
			state.msgInputCond.L.Unlock()
			break
		}

		if is_option_activated {
			state.isOptionCond = true
			is_option_activated = false
			state.optionCond.Broadcast()
			state.msgInputCond.L.Unlock()
			continue
		}

		if state.is_chat_transfer_signal {
			state.isChatLogCond = true
			state.chatLogCond.Broadcast()
			state.msgInputCond.L.Unlock()
			continue
		}

		message = message[prefix_length:]

		// if err != nil {
		// 	panic(err)
		// }
		if strings.Compare(message, "quit") == 0 {
			messageListener <- "quit"
			state.killSwitch = true
			state.isOptionCond = true
			state.optionCond.Broadcast()
			state.msgInputCond.L.Unlock()
			break
		} else if strings.Compare(message, "/c up") == 0 {
			messageListener <- "/c up"
		} else if strings.Compare(message, "/c down") == 0 {
			messageListener <- "/c down"
		} else {
			// 입력받은 메세지를 서버로 전송해야함.
			send_request.SendTransferMessage(chatRoomID, message)
		}
		state.isChatLogCond = true
		state.chatLogCond.Broadcast()
		state.msgInputCond.L.Unlock()
	}

	state.wg.Done()
}
