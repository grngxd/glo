package hotkey

import (
	"syscall"
	"unsafe"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	registerHotKey = user32.NewProc("RegisterHotKey")
	getMessage     = user32.NewProc("GetMessageW")
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
		return err
	}
	return nil
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
		getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if msg.message == WM_HOTKEY {
			callback(int(msg.wParam))
		}
	}
}
