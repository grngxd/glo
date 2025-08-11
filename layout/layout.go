package layout

import "lattice/window"

func TileWindows(windows []*window.Window, screenWidth, screenHeight, padding int) {
    if len(windows) == 0 {
        return
    }

    innerW := screenWidth - padding*2
    innerH := screenHeight - padding*2

    masterFrac := 0.6
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
