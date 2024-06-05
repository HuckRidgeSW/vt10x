// Additions to vt10x for use in hucksh.
//
// In a separate file to help avoid merge conflicts with future upstream
// changes.

package vt10x

import (
	"fmt"
	"image/color"
	"time"
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

// Used for 6x6x6 color palette. Similar to "web-safe" colors, but different.
var mod6color = []uint8{
	0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff,
}

// Used for gray-scale
var mod24color = []uint8{
	// 8 - 118 in steps of 10
	0x08, 0x12, 0x1c, 0x26, 0x30, 0x3a, 0x44, 0x4e, 0x58, 0x62, 0x6c, 0x76,
	// 128 - 238, in steps of 10
	0x80, 0x8a, 0x94, 0x9e, 0xa8, 0xb2, 0xbc, 0xc6, 0xd0, 0xda, 0xe4, 0xee,
}

// ToRealColor converts c to image/color.Color. For the first 15, it uses a
// static lookup table based on the iTerm "Light Background" color profile.
// For the rest, it uses a standard map from Wikipedia
// (https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit).
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
	}

	if 16 <= c && c <= 231 {
		// 16-231: 6x6x6 cube, 216 colors: 16 + 36r + 6g + b, 0 <= r, g, b <= 5
		c -= 16
		b := mod6color[c%6]
		c /= 6
		g := mod6color[c%6]
		c /= 6
		r := mod6color[c%6]
		return color.NRGBA{A: 0xff, R: r, G: g, B: b}

	}

	gray := mod24color[(c%256)-232]
	return color.NRGBA{A: 0xff, R: gray, G: gray, B: gray}
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
func (t *State) GlobalCell(col, row int) (ch rune, mode int16, fg, bg Color, lt time.Time) {
	g, lt := t.GlobalGlyph(col, row)
	return g.c, g.mode, g.fg, g.bg, lt
}

// GlobalGlyph returns the given glyph, including t.history.
func (t *State) GlobalGlyph(col, row int) (glyph, time.Time) {
	var l line
	if t.RecordHistory {
		hl := len(t.history)
		if row < hl {
			l = t.history[row]
		} else {
			l = t.lines[row-hl]
		}
	} else {
		l = t.lines[row]
	}
	return l.g[col], l.t
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

// LastLine returns the "global" row number of the last line with non-blank
// input. All history lines, if any, are assumed to have non-blank input. For
// a command with no output, LastLine will return -1.
func (t *State) LastLine() int {
	// This is correct even if t.getLastLine() == -1
	return len(t.history) + t.getLastLine()
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
