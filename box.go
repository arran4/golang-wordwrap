package wordwrap

import (
	"errors"
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
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
	// Len the length of the buffer represented by the box
	Len() int
	// TextValue extracts the text value
	TextValue() string
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
	// Push puts a box back on to the cache stack
	Push(box ...Box)
	// Pos text pos
	Pos() int
	// Unshift basically is Push but to the start
	Unshift(b ...Box)
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
	cacheQueue     []Box
}

// Ensures that SimpleBoxer fits model
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

// Pos current parser position.
func (sb *SimpleBoxer) Pos() int {
	r := sb.n
	for _, e := range sb.cacheQueue {
		r -= e.Len()
	}
	return r
}

// HasNext unprocessed bytes exist
func (sb *SimpleBoxer) HasNext() bool {
	return len(sb.cacheQueue) > 0 || sb.n < len(sb.text)
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

// Push puts a box back on to the cache stack
func (sb *SimpleBoxer) Push(box ...Box) {
	sb.cacheQueue = append(sb.cacheQueue, box...)
}

// Unshift basically is Push but to the start
func (sb *SimpleBoxer) Unshift(box ...Box) {
	sb.cacheQueue = append(append(make([]Box, 0, len(box)+len(sb.cacheQueue)), box...), sb.cacheQueue...)
}

// Next gets the next word in a Box
func (sb *SimpleBoxer) Next() (Box, int, error) {
	if len(sb.cacheQueue) > 0 {
		cb := sb.cacheQueue[0]
		sb.cacheQueue = sb.cacheQueue[1:]
		return cb, 0, nil
	}
	if len(sb.text) == 0 {
		return nil, 0, nil
	}
	n, rs, rmode := sb.Grabber(sb.text[sb.n:])
	sb.n += n
	var b Box
	drawer := sb.fontDrawer
	switch rmode {
	case RNIL:
		return nil, n, nil
	case RSimpleBox, RCRLF:
		t := string(rs)
		var err error
		b, err = NewSimpleTextBox(drawer, t)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, fmt.Errorf("unknown rmode %d", rmode)
	}
	switch rmode {
	case RCRLF:
		b = &LineBreakBox{
			Box: b,
		}
	}
	for _, option := range sb.postBoxOptions {
		option(b)
	}
	return b, n, nil
}

// Matches objects
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

// SimpleTextBox represents an indivisible series of characters.
type SimpleTextBox struct {
	Contents string
	Bounds   fixed.Rectangle26_6
	drawer   *font.Drawer
	Advance  fixed.Int26_6
	Metrics  font.Metrics
	boxBox   bool
}

// NewSimpleTextBox constructor
func NewSimpleTextBox(drawer *font.Drawer, t string) (Box, error) {
	if drawer == nil {
		return nil, errors.New("font drawer not provided")
	}
	ttb, a := drawer.BoundString(t)
	b := &SimpleTextBox{
		drawer:   drawer,
		Contents: t,
		Bounds:   ttb,
		Advance:  a,
		Metrics:  drawer.Face.Metrics(),
	}
	return b, nil
}

// TextValue stored value of the box
func (sb *SimpleTextBox) TextValue() string {
	return sb.Contents
}

// Len is the string length of the contents of the box
func (sb *SimpleTextBox) Len() int {
	return len(sb.Contents)
}

// FontDrawer font used
func (sb *SimpleTextBox) FontDrawer() *font.Drawer {
	return sb.drawer
}

// turnOnBox draws a box around the box
func (sb *SimpleTextBox) turnOnBox() {
	sb.boxBox = true
}

// AdvanceRect width of text
func (sb *SimpleTextBox) AdvanceRect() fixed.Int26_6 {
	return sb.Advance
}

// MetricsRect all other font details of text
func (sb *SimpleTextBox) MetricsRect() font.Metrics {
	return sb.Metrics
}

// Whitespace if this is a white space or not
func (sb *SimpleTextBox) Whitespace() bool {
	return sb.Contents == "" || unicode.IsSpace(rune(sb.Contents[0]))
}

// DrawBox renders object
func (sb *SimpleTextBox) DrawBox(i Image, y fixed.Int26_6) {
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

// LineBreakBox represents a natural or an effective line break
type LineBreakBox struct {
	Box
}

// DrawBox renders object
func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6) {}

// AdvanceRect width of text
func (sb *LineBreakBox) AdvanceRect() fixed.Int26_6 {
	return fixed.Int26_6(0)
}

// PageBreakBox represents a natural or an effective page break
type PageBreakBox struct {
	Box
}

// ImageBox is a box that contains an image
type ImageBox struct {
	I image.Image
}

// Interface enforcement
var _ Box = (*ImageBox)(nil)

// AdvanceRect width of text
func (ib *ImageBox) AdvanceRect() fixed.Int26_6 {
	return fixed.I(ib.I.Bounds().Dx())
}

// MetricsRect all other font details of text
func (ib *ImageBox) MetricsRect() font.Metrics {
	return font.Metrics{
		Height: fixed.I(ib.I.Bounds().Dy()),
		Ascent: fixed.I(ib.I.Bounds().Dy()),
	}
}

// Whitespace if this is a white space or not
func (ib *ImageBox) Whitespace() bool {
	return false
}

// DrawBox renders object
func (ib *ImageBox) DrawBox(i Image, y fixed.Int26_6) {
	draw.Draw(i, i.Bounds(), ib.I, ib.I.Bounds().Min, draw.Over)
}

// FontDrawer font used
func (ib *ImageBox) FontDrawer() *font.Drawer {
	return nil
}

// Len the length of the buffer represented by the box
func (ib *ImageBox) Len() int {
	return 0
}

// TextValue returns the text suppressed by the line break (probably a white space including a \r\n)
func (ib *ImageBox) TextValue() string {
	return ""
}

// NewImageBox constructs a new ImageBox
func NewImageBox(i image.Image) *ImageBox {
	return &ImageBox{
		I: i,
	}
}
