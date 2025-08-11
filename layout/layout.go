package layout

import "glo/window"

func TileWindows(windows []*window.Window, screenWidth, screenHeight, padding int, masterFrac float64) {
	if len(windows) == 0 {
		return
	}

	if masterFrac < 0.1 {
		masterFrac = 0.1
	} else if masterFrac > 0.9 {
		masterFrac = 0.9
	}

	innerW := screenWidth - padding*2
	innerH := screenHeight - padding*2

	masterW := int(float64(innerW) * masterFrac)
	if len(windows) == 1 {
		masterW = innerW
	}

	master := windows[0]
	master.Restore()
	master.SetRect(
		padding,
		padding,
		masterW,
		innerH,
	)

	if len(windows) == 1 {
		return
	}

	stackX := padding + masterW
	stackW := innerW - masterW

	stackCount := len(windows) - 1
	eachH := innerH / stackCount

	for i, w := range windows[1:] {
		y := padding + i*eachH
		h := eachH
		if i == stackCount-1 {
			h = innerH - eachH*(stackCount-1)
		}
		w.Restore()
		w.SetRect(stackX, y, stackW, h)
	}
}

func TileWindowsInRect(windows []*window.Window, x, y, width, height, padding int, masterFrac float64) {
	if len(windows) == 0 {
		return
	}
	if masterFrac < 0.1 {
		masterFrac = 0.1
	} else if masterFrac > 0.9 {
		masterFrac = 0.9
	}
	innerW := width - padding*2
	innerH := height - padding*2
	if innerW <= 0 || innerH <= 0 {
		return
	}
	masterW := int(float64(innerW) * masterFrac)
	if len(windows) == 1 {
		masterW = innerW
	}

	master := windows[0]
	master.Restore()
	master.SetRect(
		x+padding,
		y+padding,
		masterW,
		innerH,
	)
	if len(windows) == 1 {
		return
	}
	stackX := x + padding + masterW
	stackW := innerW - masterW
	stackCount := len(windows) - 1
	eachH := innerH / stackCount
	for i, w := range windows[1:] {
		yy := y + padding + i*eachH
		h := eachH
		if i == stackCount-1 {
			h = innerH - eachH*(stackCount-1)
		}
		w.Restore()
		w.SetRect(stackX, yy, stackW, h)
	}
}
