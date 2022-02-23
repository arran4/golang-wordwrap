package wordwrap

import (
	"errors"
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"unicode"
)

// Box is a representation of a non-divisible unit (can be nested)
type Box interface {
	// AdvanceRect width of text
	AdvanceRect() fixed.Int26_6
	// MetricsRect all other font details of text
	MetricsRect() font.Metrics
	// Whitespace if this is a white space or not
	Whitespace() bool
	// DrawBox renders object
	DrawBox(i Image, y fixed.Int26_6)
	// FontDrawer font used
	FontDrawer() *font.Drawer
}

// Boxer is the tokenizer that splits the line into it's literal components
type Boxer interface {
	// Next gets the next word in a Box
	Next() (Box, int, error)
	// SetFontDrawer Changes the default font
	SetFontDrawer(face *font.Drawer)
	// FontDrawer encapsulates default fonts and more
	FontDrawer() *font.Drawer
	// Back goes back i spaces (ie unreads)
	Back(i int)
	// HasNext if there are any unprocessed runes
	HasNext() bool
}

// IsCR Is a carriage return
func IsCR(r rune) bool {
	return r == '\r'
}

// IsLF is a line feed
func IsLF(r rune) bool {
	return r == '\n'
}

// SimpleBoxer simple tokenizer basically determines if something unicode.IsSpace or is a new line, or is text and tells
// the calling Folder that. Putting the elements in the correct Box.
type SimpleBoxer struct {
	postBoxOptions []func(Box)
	text           []rune
	n              int
	fontDrawer     *font.Drawer
	Grabber        func(text []rune) (int, []rune, int)
}

var _ Boxer = (*SimpleBoxer)(nil)

// NewSimpleBoxer simple tokenizer basically determines if something unicode.IsSpace or is a new line, or is text and tells
// the calling Folder that. Putting the elements in the correct Box.
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

// HasNext unprocessed bytes exist
func (sb *SimpleBoxer) HasNext() bool {
	return sb.n < len(sb.text)
}

// SetFontDrawer Changes the default font
func (sb *SimpleBoxer) SetFontDrawer(face *font.Drawer) {
	sb.fontDrawer = face
}

// Back goes back i spaces (ie unreads)
func (sb *SimpleBoxer) Back(i int) {
	sb.n -= i
}

// FontDrawer encapsulates default fonts and more
func (sb *SimpleBoxer) FontDrawer() *font.Drawer {
	return sb.fontDrawer
}

// Next gets the next word in a Box
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

// SimpleBoxerGrab Consumer of characters until change. Could be made to conform to strings.Scanner
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

// IsSpaceButNotCRLF Because spaces are different to CR and LF for word wrapping
func IsSpaceButNotCRLF(r rune) bool {
	return unicode.IsSpace(r) && r != '\n' && r != '\r'
}

// Once counts matches
func Once(f func(r rune) bool) func(rune) bool {
	c := 0
	return func(r rune) bool {
		c++
		return c == 1 && f(r)
	}
}

// SimpleBox represents an indivisible series of characters.
type SimpleBox struct {
	Contents string
	Bounds   fixed.Rectangle26_6
	drawer   *font.Drawer
	Advance  fixed.Int26_6
	Metrics  font.Metrics
	boxBox   bool
}

// FontDrawer font used
func (sb *SimpleBox) FontDrawer() *font.Drawer {
	return sb.drawer
}

// turnOnBox draws a box around the box
func (sb *SimpleBox) turnOnBox() {
	sb.boxBox = true
}

// AdvanceRect width of text
func (sb *SimpleBox) AdvanceRect() fixed.Int26_6 {
	return sb.Advance
}

// MetricsRect all other font details of text
func (sb *SimpleBox) MetricsRect() font.Metrics {
	return sb.Metrics
}

// Whitespace if this is a white space or not
func (sb *SimpleBox) Whitespace() bool {
	return sb.Contents == "" || unicode.IsSpace(rune(sb.Contents[0]))
}

// DrawBox renders object
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

// LineBreakBox represents a natural or a effective line break
type LineBreakBox struct {
	fontDrawer *font.Drawer
}

// FontDrawer font used
func (sb *LineBreakBox) FontDrawer() *font.Drawer {
	return sb.fontDrawer
}

// DrawBox renders object
func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6) {}

// AdvanceRect width of text
func (sb *LineBreakBox) AdvanceRect() fixed.Int26_6 {
	return fixed.Int26_6(0)
}

// MetricsRect all other font details of text
func (sb *LineBreakBox) MetricsRect() font.Metrics {
	return sb.fontDrawer.Face.Metrics()
}

// Whitespace if this is a white space or not
func (sb *LineBreakBox) Whitespace() bool {
	return true
}
