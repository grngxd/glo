package window

import (
	"fmt"
	"syscall"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")

type Window struct {
	hwnd uintptr

	width  int
	height int

	x int
	y int
}

func New(hwnd uintptr) *Window {
	w := &Window{hwnd: hwnd}

	if err := w.updateRect(); err != nil {
		panic(fmt.Errorf("failed to create window: %v", err))
	}

	return w
}

func (w *Window) Hwnd() uintptr {
	return w.hwnd
}

func (w *Window) updateRect() error {
	gwr := user32.NewProc("GetWindowRect")

	var rect struct {
		Left   int32
		Top    int32
		Right  int32
		Bottom int32
	}

	r, _, _ := gwr.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
	if r == 0 {
		return fmt.Errorf("GetWindowRect failed")
	}

	w.x = int(rect.Left)
	w.y = int(rect.Top)
	w.width = int(rect.Right - rect.Left)
	w.height = int(rect.Bottom - rect.Top)

	return nil
}

func (w *Window) GetPosition() (int, int) {
	err := w.updateRect()
	if err != nil {
		panic(fmt.Errorf("failed to get window position: %v", err))
	}

	return w.x, w.y
}

func (w *Window) GetSize() (int, int) {
	err := w.updateRect()
	if err != nil {
		panic(fmt.Errorf("failed to get window size: %v", err))
	}

	return w.width, w.height
}

func (w *Window) MoveTo(x, y int) error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	move := user32.NewProc("MoveWindow")
	r, _, _ := move.Call(w.hwnd, uintptr(x), uintptr(y), uintptr(w.width), uintptr(w.height), 1)
	if r == 0 {
		return fmt.Errorf("MoveWindow failed")
	}

	w.x = x
	w.y = y

	return nil
}

func (w *Window) MoveDelta(dx, dy int) error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	move := user32.NewProc("MoveWindow")
	r, _, _ := move.Call(w.hwnd, uintptr(w.x+dx), uintptr(w.y+dy), uintptr(w.width), uintptr(w.height), 1)
	if r == 0 {
		return fmt.Errorf("MoveWindow failed")
	}

	w.x += dx
	w.y += dy

	return nil
}

func (w *Window) Resize(width, height int) error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	resize := user32.NewProc("MoveWindow")
	r, _, _ := resize.Call(w.hwnd, uintptr(w.x), uintptr(w.y), uintptr(width), uintptr(height), 1)
	if r == 0 {
		return fmt.Errorf("MoveWindow failed")
	}

	w.width = width
	w.height = height

	return nil
}

func (w *Window) ResizeDelta(dw, dh int) error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	resize := user32.NewProc("MoveWindow")
	r, _, _ := resize.Call(w.hwnd, uintptr(w.x), uintptr(w.y), uintptr(w.width+dw), uintptr(w.height+dh), 1)
	if r == 0 {
		return fmt.Errorf("MoveWindow failed")
	}

	w.width += dw
	w.height += dh

	return nil
}

func (w *Window) GetRect() (int, int, int, int) {
	err := w.updateRect()
	if err != nil {
		panic(fmt.Errorf("failed to get window rect: %v", err))
	}

	return w.x, w.y, w.width, w.height
}

func (w *Window) showWindow(nCmdShow int) error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	maximise := user32.NewProc("ShowWindow")
	r, _, _ := maximise.Call(w.hwnd, uintptr(nCmdShow))
	if r == 0 {
		return fmt.Errorf("ShowWindow failed")
	}

	return nil
}

func (w *Window) Minimise() error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	return w.showWindow(2) // SW_SHOWMINIMIZED
}

func (w *Window) Maximise() error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	return w.showWindow(3) // SW_SHOWMAXIMIZED / SW_MAXIMIZE
}

func (w *Window) Restore() error {
	err := w.updateRect()
	if err != nil {
		return fmt.Errorf("failed to get window position: %v", err)
	}

	return w.showWindow(1) // SW_SHOWNORMAL / SW_NORMAL
}
