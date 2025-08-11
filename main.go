package main

import (
	"fmt"
	"lattice/window"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println(window.UsableScreenDimensions())
	screenWidth, screenHeight := window.UsableScreenDimensions()

	user32 := syscall.NewLazyDLL("user32.dll")
	gfw := user32.NewProc("GetForegroundWindow")

	var windows []*window.Window

	padding := 25

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	running := true

	for running {
		select {
		case <-exitChan:
			fmt.Println("cleaning up")

			for _, w := range windows {
				w.Restore()
			}

			running = false

		default:
			hwnd, _, _ := gfw.Call()
			if hwnd == 0 {
				time.Sleep(10 * time.Millisecond)
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

				w.Restore()
				w.SetRect(padding, padding, screenWidth-padding*2, screenHeight-padding*2)
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
