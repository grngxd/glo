package layout

import (
	"lattice/window"
)

// Mode defines the active tiling algorithm.
type Mode int

const (
	ModeMasterStack Mode = iota
	ModeEvenVertical
	ModeEvenHorizontal
	ModeMonocle
)

func (m Mode) String() string {
	switch m {
	case ModeMasterStack:
		return "master-stack"
	case ModeEvenVertical:
		return "even-vertical"
	case ModeEvenHorizontal:
		return "even-horizontal"
	case ModeMonocle:
		return "monocle"
	default:
		return "unknown"
	}
}

// MasterStack options.
type MasterStack struct{ MasterRatio float64 }

func (m *MasterStack) clamp() {
	if m.MasterRatio < 0.2 {
		m.MasterRatio = 0.2
	} else if m.MasterRatio > 0.8 {
		m.MasterRatio = 0.8
	}
}

// TileMasterStack tiles using master + stack.
func (m *MasterStack) TileMasterStack(handles []uintptr, work window.Rect) {
	n := len(handles)
	if n == 0 {
		return
	}
	m.clamp()
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	if n == 1 {
		_ = window.MoveResize(handles[0], work.Left, work.Top, w, h)
		return
	}
	masterW := int32(float64(w) * m.MasterRatio)
	stackW := w - masterW
	_ = window.MoveResize(handles[0], work.Left, work.Top, masterW, h)
	stackCount := n - 1
	cellH := h / int32(stackCount)
	y := work.Top
	for i := 1; i < n; i++ {
		height := cellH
		if i == n-1 {
			height = (work.Top + h) - y
		}
		_ = window.MoveResize(handles[i], work.Left+masterW, y, stackW, height)
		y += cellH
	}
}

// TileEvenVertical splits the work area into equal vertical columns.
func TileEvenVertical(handles []uintptr, work window.Rect) {
	n := len(handles)
	if n == 0 {
		return
	}
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	colW := w / int32(n)
	x := work.Left
	for i, hnd := range handles {
		width := colW
		if i == n-1 {
			width = (work.Left + w) - x
		}
		_ = window.MoveResize(hnd, x, work.Top, width, h)
		x += colW
	}
}

// TileEvenHorizontal splits the work area into equal horizontal rows.
func TileEvenHorizontal(handles []uintptr, work window.Rect) {
	n := len(handles)
	if n == 0 {
		return
	}
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	rowH := h / int32(n)
	y := work.Top
	for i, hnd := range handles {
		height := rowH
		if i == n-1 {
			height = (work.Top + h) - y
		}
		_ = window.MoveResize(hnd, work.Left, y, w, height)
		y += rowH
	}
}

// TileMonocle makes the first window full-screen (work area) and hides geometry differences for others (stacked underneath).
func TileMonocle(handles []uintptr, work window.Rect) {
	if len(handles) == 0 {
		return
	}
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	_ = window.MoveResize(handles[0], work.Left, work.Top, w, h)
	// Optionally also resize others so if promoted they already fit.
	for i := 1; i < len(handles); i++ {
		_ = window.MoveResize(handles[i], work.Left, work.Top, w, h)
	}
}

// Tile dispatches based on mode.
func Tile(mode Mode, ms *MasterStack, handles []uintptr, work window.Rect) {
	switch mode {
	case ModeMasterStack:
		ms.TileMasterStack(handles, work)
	case ModeEvenVertical:
		TileEvenVertical(handles, work)
	case ModeEvenHorizontal:
		TileEvenHorizontal(handles, work)
	case ModeMonocle:
		TileMonocle(handles, work)
	}
}
