package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
)

const (
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008

	VK_LEFT      = 0x25
	VK_UP        = 0x26
	VK_RIGHT     = 0x27
	VK_DOWN      = 0x28
	VK_RETURN    = 0x0D
	VK_BACK      = 0x08
	VK_OEM_PLUS  = 0xBB
	VK_OEM_MINUS = 0xBD
	VK_T         = 0x54
)

func Register(id int, modifiers uint, vk uint) error {
	r, _, err := procRegisterHotKey.Call(0, uintptr(id), uintptr(modifiers), uintptr(vk))
	if r == 0 {
		return fmt.Errorf("RegisterHotKey failed: %v", err)
	}
	return nil
}

func Unregister(id int) { procUnregisterHotKey.Call(0, uintptr(id)) }

type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       struct{ X, Y int32 }
	LPrivate uint32
}

const WM_HOTKEY = 0x0312

func Loop(onHotkey func(id int)) error {
	var msg MSG
	for {
		r, _, e := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int32(r) {
		case -1:
			return fmt.Errorf("GetMessage failed: %v", e)
		case 0:
			return nil
		}
		if msg.Message == WM_HOTKEY {
			onHotkey(int(msg.WParam))
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
