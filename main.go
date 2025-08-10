package main

import (
	"fmt"
	"lattice/window"
	"syscall"
	"time"
)

func main() {
	user32 := syscall.NewLazyDLL("user32.dll")

	gfw := user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := gfw.Call()
	if hwnd == 0 {
		panic("failed to get foreground window")
	}

	w := window.New(hwnd)
	fmt.Printf("Window Handle: %d\n", w.Hwnd())

	time.Sleep(3 * time.Second)

	w.Restore()
}
