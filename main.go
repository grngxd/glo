package main

import (
	"fmt"
	"lattice/hotkey"
	"lattice/layout"
	"lattice/window"
	"log"
	"slices"
)

const (
	HK_ADD_WINDOW = iota + 1
	HK_REMOVE_WINDOW
	HK_MASTER_INC
	HK_MASTER_DEC
	HK_ROTATE_NEXT
	HK_PROMOTE
	HK_CYCLE_LAYOUT
)

func main() {
	fmt.Println("Master-stack tiler:\n  Ctrl+Alt+Enter add window  | Ctrl+Alt+Backspace remove window\n  Ctrl+Alt+= / Ctrl+Alt+- adjust master size\n  Ctrl+Alt+Right rotate | Ctrl+Alt+Left promote focused to master\n  Ctrl+Alt+T cycle layout mode\n  Ctrl+C to exit.")

	mods := hotkey.MOD_CONTROL | hotkey.MOD_ALT
	must := func(id int, key uint) {
		if err := hotkey.Register(id, uint(mods), key); err != nil {
			log.Fatal(err)
		}
	}
	must(HK_ADD_WINDOW, hotkey.VK_RETURN)
	must(HK_REMOVE_WINDOW, hotkey.VK_BACK)
	must(HK_MASTER_INC, hotkey.VK_OEM_PLUS)
	must(HK_MASTER_DEC, hotkey.VK_OEM_MINUS)
	must(HK_ROTATE_NEXT, hotkey.VK_RIGHT)
	must(HK_PROMOTE, hotkey.VK_LEFT)
	must(HK_CYCLE_LAYOUT, hotkey.VK_T)
	defer func() {
		hotkey.Unregister(HK_ADD_WINDOW)
		hotkey.Unregister(HK_REMOVE_WINDOW)
		hotkey.Unregister(HK_MASTER_INC)
		hotkey.Unregister(HK_MASTER_DEC)
		hotkey.Unregister(HK_ROTATE_NEXT)
		hotkey.Unregister(HK_PROMOTE)
		hotkey.Unregister(HK_CYCLE_LAYOUT)
	}()

	var handles []uintptr
	mode := layout.ModeMasterStack
	ms := &layout.MasterStack{MasterRatio: 0.6}

	retile := func(focus uintptr) {
		if len(handles) == 0 {
			return
		}
		cleaned := handles[:0]
		for _, h := range handles {
			if _, err := window.GetRect(h); err == nil {
				cleaned = append(cleaned, h)
			}
		}
		handles = cleaned
		if len(handles) == 0 {
			return
		}
		base := handles[0]
		if focus != 0 {
			base = focus
		}
		work, err := window.WorkArea(base)
		if err != nil {
			return
		}
		layout.Tile(mode, ms, handles, work)
	}

	contains := func(h uintptr) bool {
		for _, x := range handles {
			if x == h {
				return true
			}
		}
		return false
	}

	handle := func(id int) {
		hwnd := window.Foreground()
		if hwnd == 0 {
			return
		}
		switch id {
		case HK_ADD_WINDOW:
			if !contains(hwnd) {
				handles = append(handles, hwnd)
			}
		case HK_REMOVE_WINDOW:
			for i, h := range handles {
				if h == hwnd {
					handles = append(handles[:i], handles[i+1:]...)
					break
				}
			}
		case HK_MASTER_INC:
			ms.MasterRatio += 0.05
		case HK_MASTER_DEC:
			ms.MasterRatio -= 0.05
		case HK_ROTATE_NEXT:
			if len(handles) > 1 {
				first := handles[0]
				handles = append(handles[1:], first)
			}
		case HK_PROMOTE:
			if len(handles) > 1 {
				if idx := slices.Index(handles, hwnd); idx > 0 {
					handles[0], handles[idx] = handles[idx], handles[0]
				}
			}
		case HK_CYCLE_LAYOUT:
			mode = (mode + 1) % 4
			fmt.Println("Layout:", mode.String())
		}
		retile(hwnd)
	}
	if err := hotkey.Loop(handle); err != nil {
		log.Fatal(err)
	}
}
