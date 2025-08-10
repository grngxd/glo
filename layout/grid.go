package layout

import "lattice/window"

type Grid struct{ Cols, Rows int }
type Cell struct{ Col, Row int }

func (g Grid) FitWindow(r window.Rect, work window.Rect) Cell {
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	cellW := w / int32(g.Cols)
	cellH := h / int32(g.Rows)
	cx := (r.Left+r.Right)/2 - work.Left
	cy := (r.Top+r.Bottom)/2 - work.Top
	col := int(cx / cellW)
	row := int(cy / cellH)
	if col < 0 {
		col = 0
	} else if col >= g.Cols {
		col = g.Cols - 1
	}
	if row < 0 {
		row = 0
	} else if row >= g.Rows {
		row = g.Rows - 1
	}
	return Cell{col, row}
}

func (g Grid) RectFor(c Cell, work window.Rect) window.Rect {
	w := work.Right - work.Left
	h := work.Bottom - work.Top
	cellW := w / int32(g.Cols)
	cellH := h / int32(g.Rows)
	left := work.Left + int32(c.Col)*cellW
	top := work.Top + int32(c.Row)*cellH
	right := left + cellW
	bottom := top + cellH
	return window.Rect{Left: left, Top: top, Right: right, Bottom: bottom}
}
