package window

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procMonitorFromWindow   = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
)

type Rect struct{ Left, Top, Right, Bottom int32 }

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor Rect
	RcWork    Rect
	DwFlags   uint32
}

const (
	MONITOR_DEFAULTTONEAREST = 0x00000002
	SWP_NOZORDER             = 0x0004
	SWP_NOACTIVATE           = 0x0010
)

func Foreground() uintptr { h, _, _ := procGetForegroundWindow.Call(); return h }

func GetRect(hwnd uintptr) (Rect, error) {
	var r Rect
	ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ok == 0 {
		return r, errors.New("GetWindowRect failed")
	}
	return r, nil
}

func WorkArea(hwnd uintptr) (Rect, error) {
	m, _, _ := procMonitorFromWindow.Call(hwnd, MONITOR_DEFAULTTONEAREST)
	if m == 0 {
		return Rect{}, errors.New("MonitorFromWindow failed")
	}
	var mi MONITORINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	ok, _, _ := procGetMonitorInfoW.Call(m, uintptr(unsafe.Pointer(&mi)))
	if ok == 0 {
		return Rect{}, errors.New("GetMonitorInfoW failed")
	}
	return mi.RcWork, nil
}

func MoveResize(hwnd uintptr, left, top, width, height int32) error {
	r, _, _ := procSetWindowPos.Call(hwnd, 0, uintptr(left), uintptr(top), uintptr(width), uintptr(height), SWP_NOZORDER|SWP_NOACTIVATE)
	if r == 0 {
		return errors.New("SetWindowPos failed")
	}
	return nil
}
