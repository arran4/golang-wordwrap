package wordwrap

import (
	"errors"
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"unicode"
)

type Box interface {
	FontRect() fixed.Rectangle26_6
	AdvanceRect() fixed.Int26_6
	MetricsRect() font.Metrics
	Whitespace() bool
	DrawBox(i Image, y fixed.Int26_6)
	FontDrawer() *font.Drawer
}

type Boxer interface {
	Next() (Box, int, error)
	SetFontDrawer(face *font.Drawer)
	FontDrawer() *font.Drawer
	Back(i int)
}

func IsCR(r rune) bool {
	return r == '\r'
}

func IsLF(r rune) bool {
	return r == '\n'
}

type SimpleBoxer struct {
	postBoxOptions []func(Box)
	text           []rune
	n              int
	fontDrawer     *font.Drawer
	Grabber        func(text []rune) (int, []rune, int)
}

func NewSimpleBoxer(text []rune, drawer *font.Drawer, options ...BoxerOption) *SimpleBoxer {
	sb := &SimpleBoxer{
		text:       text,
		n:          0,
		fontDrawer: drawer,
		Grabber:    SimpleBoxerGrab,
	}
	for _, option := range options {
		option.ApplyBoxConfig(sb)
	}
	return sb
}

func (sb *SimpleBoxer) SetFontDrawer(face *font.Drawer) {
	sb.fontDrawer = face
}

func (sb *SimpleBoxer) Back(i int) {
	sb.n -= i
}

func (sb *SimpleBoxer) FontDrawer() *font.Drawer {
	return sb.fontDrawer
}

func (sb *SimpleBoxer) Next() (Box, int, error) {
	if len(sb.text) == 0 {
		return nil, 0, nil
	}
	n, rs, rmode := sb.Grabber(sb.text[sb.n:])
	sb.n += n
	var b Box
	switch rmode {
	case RNIL:
		return nil, n, nil
	case RCRLF:
		b = &LineBreakBox{
			fontDrawer: sb.fontDrawer,
		}
	case RSimpleBox:
		t := string(rs)
		if sb.fontDrawer == nil {
			return nil, 0, errors.New("font drawer not provided")
		}
		ttb, a := sb.fontDrawer.BoundString(t)
		b = &SimpleBox{
			drawer:   sb.fontDrawer,
			Contents: t,
			Bounds:   ttb,
			Advance:  a,
			Metrics:  sb.fontDrawer.Face.Metrics(),
		}
	default:
		return nil, 0, fmt.Errorf("unknown rmode %d", rmode)
	}
	for _, option := range sb.postBoxOptions {
		option(b)
	}
	return b, n, nil
}

const (
	RSimpleBox = iota
	RCRLF
	RNIL
)

func SimpleBoxerGrab(text []rune) (int, []rune, int) {
	n := 0
	rmode := RNIL
	var mode func(rune) bool
	for _, r := range text {
		if mode == nil {
			if IsCR(r) {
				mode = Once(func(r rune) bool {
					if IsLF(r) {
						rmode = RCRLF
						return true
					}
					return false
				})
				n++
				continue
			} else if IsLF(r) {
				rmode = RCRLF
				n++
				break
			} else if !unicode.IsPrint(r) {
				continue
			} else if IsSpaceButNotCRLF(r) {
				mode = IsSpaceButNotCRLF
				rmode = RSimpleBox
			} else {
				rmode = RSimpleBox
				mode = func(r rune) bool {
					return !unicode.IsSpace(r)
				}
			}
		}
		if !mode(r) {
			break
		}
		n++
	}
	return n, text[:n], rmode
}

func IsSpaceButNotCRLF(r rune) bool {
	return unicode.IsSpace(r) && r != '\n' && r != '\r'
}

func Once(f func(r rune) bool) func(rune) bool {
	c := 0
	return func(r rune) bool {
		c++
		return c == 1 && f(r)
	}
}

type SimpleBox struct {
	Contents string
	Bounds   fixed.Rectangle26_6
	drawer   *font.Drawer
	Advance  fixed.Int26_6
	Metrics  font.Metrics
	boxBox   bool
}

func (sb *SimpleBox) FontDrawer() *font.Drawer {
	return sb.drawer
}

func (sb *SimpleBox) turnOnBox() {
	sb.boxBox = true
}

func (sb *SimpleBox) AdvanceRect() fixed.Int26_6 {
	return sb.Advance
}

func (sb *SimpleBox) MetricsRect() font.Metrics {
	return sb.Metrics
}

func (sb *SimpleBox) FontRect() fixed.Rectangle26_6 {
	return sb.Bounds
}

func (sb *SimpleBox) Whitespace() bool {
	return sb.Contents == "" || unicode.IsSpace(rune(sb.Contents[0]))
}

func (sb *SimpleBox) DrawBox(i Image, y fixed.Int26_6) {
	sb.drawer.Dst = i
	b := i.Bounds()
	sb.drawer.Dot = fixed.Point26_6{
		X: fixed.I(b.Min.X),
		Y: fixed.I(b.Min.Y) + y,
	}
	sb.drawer.DrawString(sb.Contents)
	if sb.boxBox {
		util.DrawBox(i, b)
	}
}

type LineBreakBox struct {
	fontDrawer *font.Drawer
}

func (sb *LineBreakBox) FontDrawer() *font.Drawer {
	return sb.fontDrawer
}

func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6) {}

func (sb *LineBreakBox) AdvanceRect() fixed.Int26_6 {
	return fixed.Int26_6(0)
}

func (sb *LineBreakBox) MetricsRect() font.Metrics {
	return sb.fontDrawer.Face.Metrics()
}

func (sb *LineBreakBox) FontRect() fixed.Rectangle26_6 {
	return fixed.Rectangle26_6{}
}

func (sb *LineBreakBox) Whitespace() bool {
	return true
}
