package main

import (
	"fmt"
	"lattice/window"
	"syscall"
	"time"
)

func main() {
	fmt.Println("ready")

	user32 := syscall.NewLazyDLL("user32.dll")

	gfw := user32.NewProc("GetForegroundWindow")

	var windows []*window.Window

	for {
		hwnd, _, _ := gfw.Call()
		if hwnd == 0 {
			continue
		}

		exists := false
		for _, w := range windows {
			if w.Hwnd() == hwnd {
				exists = true
				break
			}
		}

		if !exists {
			w := window.New(hwnd)
			windows = append(windows, w)
			fmt.Printf("%v: %v\n", len(windows), w.Hwnd())
		}

		time.Sleep(10 * time.Millisecond)
	}
}
