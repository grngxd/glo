package main

import (
	"fmt"
	"lattice/hotkey"
	"lattice/layout"
	"lattice/window"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println(`
gloWM - keybinds
Win+Shift+O: toggle tiling
`)
	screenWidth, screenHeight := window.UsableScreenDimensions()

	user32 := syscall.NewLazyDLL("user32.dll")
	gfw := user32.NewProc("GetForegroundWindow")

	var windows []*window.Window

	padding := 25

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	running := true

	tilingActive := false

	go func() {
		hotkey.RegisterGlobalHotkey(1, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0x4F) // Win+Shift+O toggle
		hotkey.ListenHotkeys(func(id int) {
			if id != 1 {
				return
			}
			tilingActive = !tilingActive
			if tilingActive {
				layout.TileWindows(windows, screenWidth, screenHeight, padding)
			} else {
				// restore original geometry
				for _, w := range windows {
					w.Restore()
					w.SetRect(w.Meta.Ox, w.Meta.Oy, w.Meta.Ow, w.Meta.Oh)
				}
			}
		})
	}()

	for running {
		select {
		case <-exitChan:
			//fmt.Println("cleaning up")

			for _, w := range windows {
				w.Restore()
				w.SetRect(w.Meta.Ox, w.Meta.Oy, w.Meta.Ow, w.Meta.Oh)
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

			if !exists && window.IsAppWindow(hwnd) {
				w := window.New(hwnd)
				windows = append(windows, w)
				//fmt.Printf("added %d hwnd=%v\n", len(windows), w.Hwnd())

				w.OnMinimize(func() {
					for i, ww := range windows {
						if ww == w {
							windows = append(windows[:i], windows[i+1:]...)
							break
						}
					}
					if tilingActive {
						layout.TileWindows(windows, screenWidth, screenHeight, padding)
					}
				})

				w.OnRestore(func() {
					present := false
					for _, ww := range windows {
						if ww == w {
							present = true
							break
						}
					}
					if !present {
						windows = append(windows, w)
					}
					if tilingActive {
						layout.TileWindows(windows, screenWidth, screenHeight, padding)
					}
				})

				w.OnClose(func() {
					for i, ww := range windows {
						if ww == w {
							windows = append(windows[:i], windows[i+1:]...)
							break
						}
					}
					if tilingActive {
						layout.TileWindows(windows, screenWidth, screenHeight, padding)
					}
				})
				if tilingActive {
					layout.TileWindows(windows, screenWidth, screenHeight, padding)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
