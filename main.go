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

	type managedWindow struct {
		win                        *window.Window
		origX, origY, origW, origH int
	}

	var windows []*managedWindow

	padding := 25

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	running := true

	for running {
		select {
		case <-exitChan:
			fmt.Println("cleaning up")

			for _, mw := range windows {
				mw.win.Restore()
				mw.win.SetRect(mw.origX, mw.origY, mw.origW, mw.origH)
			}

			running = false

		default:
			hwnd, _, _ := gfw.Call()
			if hwnd == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			exists := false
			for _, mw := range windows {
				if mw.win.Hwnd() == hwnd {
					exists = true
					break
				}
			}

			if !exists {
				w := window.New(hwnd)
				x, y, w0, h0 := w.GetRect()
				windows = append(windows, &managedWindow{win: w, origX: x, origY: y, origW: w0, origH: h0})

				fmt.Printf("%v: %v\n", len(windows), w.Hwnd())

				w.Restore()
				w.SetRect(padding, padding, screenWidth-padding*2, screenHeight-padding*2)
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
