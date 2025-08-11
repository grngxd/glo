package main

import (
	"flag"
	"fmt"
	"lattice/hotkey"
	"lattice/layout"
	"lattice/window"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	paddingFlag := flag.Int("padding", 30, "outer padding in pixels")
	masterFlag := flag.Float64("master", 0.6, "master area fraction (0.1-0.9)")
	flag.Parse()

	fmt.Println("gloWM - keybinds\nWin+Shift+O: toggle tiling\nWin+Shift+=: grow master\nWin+Shift+-: shrink master\nWin+Shift+.: rotate master\nWin+Shift+Q: quit")
	screenWidth, screenHeight := window.UsableScreenDimensions()

	user32 := syscall.NewLazyDLL("user32.dll")
	gfw := user32.NewProc("GetForegroundWindow")

	var windows []*window.Window
	var mu sync.RWMutex // protects windows slice
	var tilingActive bool

	padding := *paddingFlag
	masterFrac := *masterFlag

	var tileMu sync.Mutex
	var pending bool
	var lastTile time.Time
	triggerTile := func(ws []*window.Window) {
		tileMu.Lock()
		defer tileMu.Unlock()
		if !tilingActive {
			return
		}
		if time.Since(lastTile) < 40*time.Millisecond {
			if !pending {
				pending = true
				go func() {
					time.Sleep(50 * time.Millisecond)
					tileMu.Lock()
					p := pending
					pending = false
					tileMu.Unlock()
					if p {
						mu.RLock()
						snapshot := append([]*window.Window(nil), windows...)
						mu.RUnlock()
						layout.TileWindows(snapshot, screenWidth, screenHeight, padding, masterFrac)
					}
				}()
			}
			return
		}
		lastTile = time.Now()
		layout.TileWindows(ws, screenWidth, screenHeight, padding, masterFrac)
	}

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	running := true

	// Start hotkey registration + message loop on a dedicated OS thread
	hotkeyEvents := make(chan int, 32)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// register hotkeys on this thread
		_ = hotkey.RegisterGlobalHotkey(1, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0x4F) // toggle
		_ = hotkey.RegisterGlobalHotkey(2, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0xBB) // grow master (+)
		_ = hotkey.RegisterGlobalHotkey(3, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0xBD) // shrink master (-)
		_ = hotkey.RegisterGlobalHotkey(4, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0xBE) // rotate master (.)
		_ = hotkey.RegisterGlobalHotkey(5, hotkey.MOD_WIN|hotkey.MOD_SHIFT, 0x51) // quit (Q)

		hotkey.ListenHotkeys(func(id int) {
			select {
			case hotkeyEvents <- id:
			default:
				// drop
			}
		})
	}()

	go func() {
		for id := range hotkeyEvents {
			mu.RLock()
			snapshot := append([]*window.Window(nil), windows...)
			mu.RUnlock()

			switch id {
			case 1:
				tilingActive = !tilingActive
				if tilingActive {
					layout.TileWindows(snapshot, screenWidth, screenHeight, padding, masterFrac)
				} else {
					for _, w := range snapshot {
						w.Restore()
						w.SetRect(w.Meta.Ox, w.Meta.Oy, w.Meta.Ow, w.Meta.Oh)
					}
				}
			case 2:
				masterFrac += 0.05
				if masterFrac > 0.9 {
					masterFrac = 0.9
				}
				triggerTile(snapshot)
			case 3:
				masterFrac -= 0.05
				if masterFrac < 0.1 {
					masterFrac = 0.1
				}
				triggerTile(snapshot)
			case 4:
				if len(snapshot) > 1 {
					mu.Lock()
					first := windows[0]
					copy(windows, windows[1:])
					windows[len(windows)-1] = first
					mu.Unlock()

					mu.RLock()
					updatedSnapshot := append([]*window.Window(nil), windows...)
					mu.RUnlock()
					triggerTile(updatedSnapshot)
				}
			case 5:
				exitChan <- os.Interrupt
			}
		}
	}()

	for running {
		select {
		case <-exitChan:
			mu.RLock()
			current := append([]*window.Window(nil), windows...)
			mu.RUnlock()
			for _, w := range current {
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
			mu.RLock()
			for _, w := range windows {
				if w.Hwnd() == hwnd {
					exists = true
					break
				}
			}
			mu.RUnlock()

			if !exists && window.IsAppWindow(hwnd) {
				w := window.New(hwnd)
				mu.Lock()
				windows = append(windows, w)
				mu.Unlock()

				w.OnMinimize(func() {
					mu.Lock()
					for i, ww := range windows {
						if ww == w {
							windows = append(windows[:i], windows[i+1:]...)
							break
						}
					}
					current := append([]*window.Window(nil), windows...)
					mu.Unlock()
					if tilingActive {
						layout.TileWindows(current, screenWidth, screenHeight, padding, masterFrac)
					}
				})

				w.OnRestore(func() {
					mu.Lock()
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
					current := append([]*window.Window(nil), windows...)
					mu.Unlock()
					if tilingActive {
						layout.TileWindows(current, screenWidth, screenHeight, padding, masterFrac)
					}
				})

				w.OnClose(func() {
					mu.Lock()
					for i, ww := range windows {
						if ww == w {
							windows = append(windows[:i], windows[i+1:]...)
							break
						}
					}
					current := append([]*window.Window(nil), windows...)
					mu.Unlock()
					if tilingActive {
						layout.TileWindows(current, screenWidth, screenHeight, padding, masterFrac)
					}
				})
				if tilingActive {
					mu.RLock()
					current := append([]*window.Window(nil), windows...)
					mu.RUnlock()
					layout.TileWindows(current, screenWidth, screenHeight, padding, masterFrac)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
