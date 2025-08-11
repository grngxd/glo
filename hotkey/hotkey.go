package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	registerHotKey   = user32.NewProc("RegisterHotKey")
	unregisterHotKey = user32.NewProc("UnregisterHotKey")
	getMessage       = user32.NewProc("GetMessageW")
	postQuitMessage  = user32.NewProc("PostQuitMessage")
)

const (
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008
	WM_HOTKEY   = 0x0312
)

func RegisterGlobalHotkey(id, modifiers, vk int) error {
	r, _, err := registerHotKey.Call(0, uintptr(id), uintptr(modifiers), uintptr(vk))
	if r == 0 {
		if err == syscall.Errno(0) {
			err = syscall.GetLastError()
		}
		fmt.Printf("[hotkey] failed to register hotkey (id=%d, mod=%d, vk=%d): %v\n", id, modifiers, vk, err)
		return err
	}
	//fmt.Printf("[hotkey] registered hotkey id=%d mod=%d vk=%d\n", id, modifiers, vk)

	return nil
}

func UnregisterGlobalHotkey(id int) error {
	r, _, err := unregisterHotKey.Call(0, uintptr(id))
	if r == 0 {
		if err == syscall.Errno(0) {
			err = syscall.GetLastError()
		}
		return err
	}
	return nil
}

func PostQuit(exitCode int) {
	postQuitMessage.Call(uintptr(exitCode))
}

func ListenHotkeys(callback func(id int)) {
	var msg struct {
		hwnd    uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}
	for {
		r, _, err := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 {
			//fmt.Println("[hotkey] message loop exiting (WM_QUIT)")
			return
		}
		if r == ^uintptr(0) { // -1
			if err == syscall.Errno(0) {
				err = syscall.GetLastError()
			}

			fmt.Printf("[hotkey] GetMessageW error: %v\n", err)
			continue
		}
		if msg.message == WM_HOTKEY {
			callback(int(msg.wParam))
		}
	}
}
