// Additions to vt10x for use in hucksh.
//
// In a separate file to help avoid merge conflicts with future upstream
// changes.

package vt10x

import (
	"fmt"
	"image/color"
)

type Glyph = glyph
type Line = line

const (
	AttrReverse   = attrReverse
	AttrUnderline = attrUnderline
	AttrBold      = attrBold
	AttrGfx       = attrGfx
	AttrItalic    = attrItalic
	AttrBlink     = attrBlink
	AttrWrap      = attrWrap
)

// ToRealColor converts c to image/color.Color, using a static lookup table
// based on the iTerm "Light Background" color profile.
func (c Color) ToRealColor() color.Color {
	switch c {
	case Black, DefaultFG:
		return color.NRGBA{A: 0xff}
	case Red:
		return color.NRGBA{A: 0xff, R: 0xc9, G: 0x1b, B: 0x00}
	case Green:
		return color.NRGBA{A: 0xff, R: 0x00, G: 0xc2, B: 0x00}
	case Yellow:
		return color.NRGBA{A: 0xff, R: 0xc7, G: 0xc4, B: 0x00}
	case Blue:
		return color.NRGBA{A: 0xff, R: 0x02, G: 0x25, B: 0xc7}
	case Magenta:
		return color.NRGBA{A: 0xff, R: 0xc9, G: 0x30, B: 0xc7}
	case Cyan:
		return color.NRGBA{A: 0xff, R: 0x00, G: 0xc5, B: 0xc7}
	case LightGrey: // aka "normal white" in iTerm
		return color.NRGBA{A: 0xff, R: 0xc7, G: 0xc7, B: 0xc7}

	case DarkGrey: // aka "bright black" in iTerm
		return color.NRGBA{A: 0xff, R: 0x67, G: 0x67, B: 0x67}
	case LightRed:
		return color.NRGBA{A: 0xff, R: 0xff, G: 0x6d, B: 0x67}
	case LightGreen:
		return color.NRGBA{A: 0xff, R: 0x5f, G: 0xf9, B: 0x67}
	case LightYellow:
		return color.NRGBA{A: 0xff, R: 0xfe, G: 0xfb, B: 0x67}
	case LightBlue:
		return color.NRGBA{A: 0xff, R: 0x68, G: 0x71, B: 0xff}
	case LightMagenta:
		return color.NRGBA{A: 0xff, R: 0xff, G: 0x76, B: 0xff}
	case LightCyan:
		return color.NRGBA{A: 0xff, R: 0x5f, G: 0xfd, B: 0xff}
	case White, DefaultBG: // aka "bright white" in iTerm
		return color.NRGBA{A: 0xff, R: 0xff, G: 0xfe, B: 0xfe}

	default:
		panic(fmt.Sprintf("Unknown color: %d", c))
	}
}

// GlobalSize returns rows and columns of state, including t.history.
func (t *State) GlobalSize() (rows int, cols int) {
	if t.RecordHistory {
		return t.rows + len(t.history), t.cols
	}
	return t.rows, t.cols
}

// HistLen return the length of t.history.
func (t *State) HistLen() int {
	return len(t.history)
}

// TotalLen returns the total number of lines, including t.history.
func (t *State) TotalLen() int {
	return len(t.history) + len(t.lines)
}

// GlobalCell returns the given glyph, includign t.history, as separate return
// values.
func (t *State) GlobalCell(x, y int) (ch rune, mode int16, fg, bg Color) {
	g := t.GlobalGlyph(x, y)
	return g.c, g.mode, g.fg, g.bg
}

// GlobalGlyph returns the given glyph, including t.history.
func (t *State) GlobalGlyph(x, y int) glyph {
	var g glyph
	if t.RecordHistory {
		hl := len(t.history)
		if y < hl {
			g = t.history[y][x]
		} else {
			g = t.lines[y-hl][x]
		}
	} else {
		g = t.lines[y][x]
	}
	return g
}

// GlobalRowDirty returns true if the given row is dirty (that is, changed
// since the last call to resetChanges), including t.history.
//
// t.history is never considered dirty.
func (t *State) GlobalRowDirty(row int) bool {
	if t.RecordHistory {
		hl := len(t.history)
		if row < hl {
			return false
		}
		row -= hl
	}
	return t.dirty[row]
}

// NewGlyph creates a new glyph from separate components.
func NewGlyph(c rune, mode int16, fg, bg Color) glyph {
	return glyph{c: c, mode: mode, fg: fg, bg: bg}
}

// C returns g.c.
func (g glyph) C() rune {
	return g.c
}

// Fields returns the fields of g as separate return values.
func (g glyph) Fields() (c rune, mode int16, fg, bg Color) {
	return g.c, g.mode, g.fg, g.bg
}

// Equal returns true if g and g2 have the same runes and are Similar.
func (g glyph) Equal(g2 glyph) bool {
	return g.c == g2.c &&
		g.Similar(g2)
}

// Similar returns true if g and g2 are "similar", i.e. have the same modes
// (ignoring "wrap") and foreground and background colors. Similar ignores
// glyph.c.
func (g glyph) Similar(g2 glyph) bool {
	return (g.mode&^attrWrap) == (g2.mode&^attrWrap) &&
		g.fg == g2.fg &&
		g.bg == g2.bg
}

// Format implements the fmt.Formatter interface.
func (g glyph) Format(f fmt.State, verb rune) {
	f.Write([]byte(
		fmt.Sprintf("{%q %x %x %x}",
			g.c, g.mode, g.fg, g.bg)))
}
