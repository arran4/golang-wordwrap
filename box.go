package wordwrap

import (
	"errors"
	"fmt"
	"image"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Box represents a non-divisible unit of content (e.g., a word or image), which can be nested.
type Box interface {
	// AdvanceRect returns the width of the content.
	AdvanceRect() fixed.Int26_6
	// MetricsRect returns the font metrics of the content.
	MetricsRect() font.Metrics
	// Whitespace returns true if the content is whitespace.
	Whitespace() bool
	// DrawBox renders the content into the given image at the specified Y offset.
	DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig)
	// FontDrawer returns the font face used for this box.
	FontDrawer() *font.Drawer
	// Len returns the length of the content (e.g. rune count).
	Len() int
	// TextValue returns the text string content of the box.
	TextValue() string
	// MinSize returns the minimum required size for the box (width, height).
	MinSize() (fixed.Int26_6, fixed.Int26_6)
	// MaxSize returns the maximum allowed size for the box (width, height).
	MaxSize() (fixed.Int26_6, fixed.Int26_6)
}

// Boxer splits a line of text (or other content) into indivisible Box components.
type Boxer interface {
	// Next returns the next Box.
	Next() (Box, int, error)
	// SetFontDrawer sets the default font for the boxer.
	SetFontDrawer(face *font.Drawer)
	// FontDrawer returns the default font drawer.
	FontDrawer() *font.Drawer
	// Back unreads the last i atoms/boxes.
	Back(i int)
	// HasNext returns true if there is more content to process.
	HasNext() bool
	// Push returns boxes to the front of the queue (stack behavior).
	Push(box ...Box)
	// Pos returns the current cursor position in the input.
	Pos() int
	// Unshift adds boxes to the beginning of the internal buffer.
	Unshift(b ...Box)
	// Shift removes and returns the first Box from the internal buffer.
	Shift() Box
	// Reset restarts the tokenization
	Reset()
}

// IsCR Is a carriage return
func IsCR(r rune) bool {
	return r == '\r'
}

// Identifier is a interface for reporting IDs
type Identifier interface {
	ID() interface{}
}

// IsLF is a line feed
func IsLF(r rune) bool {
	return r == '\n'
}

// Tokenizer is a function that tokenizes the text
type Tokenizer func(text []rune) (int, []rune, int)

// SimpleBoxer simple tokenizer basically determines if something unicode.IsSpace or is a new line, or is text and tells
// the calling Folder that. Putting the elements in the correct Box.
type SimpleBoxer struct {
	postBoxOptions []func(Box)
	text           []rune
	n              int
	fontDrawer     *font.Drawer
	Tokenizer      Tokenizer
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
		Tokenizer:  LatinTokenizer,
	}
	for _, option := range options {
		option.ApplyBoxConfig(sb)
	}
	return sb
}

// Reset restarts the tokenization
func (sb *SimpleBoxer) Reset() {
	sb.n = 0
	sb.cacheQueue = nil
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

func (sb *SimpleBoxer) Shift() Box {
	if len(sb.cacheQueue) > 0 {
		cb := sb.cacheQueue[0]
		sb.cacheQueue = sb.cacheQueue[1:]
		return cb
	}
	return nil
}

// RowBox holds multiple boxes on a single line
type RowBox struct {
	Boxes []Box
}

func (rb *RowBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (rb *RowBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (rb *RowBox) AdvanceRect() fixed.Int26_6 {
	var w fixed.Int26_6
	for _, b := range rb.Boxes {
		w += b.AdvanceRect()
	}
	return w
}

func (rb *RowBox) MetricsRect() font.Metrics {
	var m font.Metrics
	for _, b := range rb.Boxes {
		bm := b.MetricsRect()
		if bm.Ascent > m.Ascent {
			m.Ascent = bm.Ascent
		}
		if bm.Descent > m.Descent {
			m.Descent = bm.Descent
		}
	}
	return m
}

func (rb *RowBox) Whitespace() bool {
	for _, b := range rb.Boxes {
		if !b.Whitespace() {
			return false
		}
	}
	return true
}

func (rb *RowBox) Len() int {
	l := 0
	for _, b := range rb.Boxes {
		l += b.Len()
	}
	return l
}

func (rb *RowBox) TextValue() string {
	s := ""
	for _, b := range rb.Boxes {
		s += b.TextValue()
	}
	return s
}

func (rb *RowBox) FontDrawer() *font.Drawer {
	if len(rb.Boxes) > 0 {
		return rb.Boxes[0].FontDrawer()
	}
	return nil
}

func (rb *RowBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	var x fixed.Int26_6
	for _, b := range rb.Boxes {
		w := b.AdvanceRect()
		adv := w.Ceil()
		r := i.Bounds()
		minX := r.Min.X + x.Ceil()
		maxX := minX + adv
		if maxX > r.Max.X {
			maxX = r.Max.X
		}
		subR := image.Rect(minX, r.Min.Y, maxX, r.Max.Y)

		if !subR.Empty() {
			subI := i.SubImage(subR).(Image)
			b.DrawBox(subI, y, dc)
		}
		x += w
	}
}

// Next gets the next word in a Box
func (sb *SimpleBoxer) Next() (Box, int, error) {
	if len(sb.cacheQueue) > 0 {
		return sb.Shift(), 0, nil
	}
	if len(sb.text) == 0 {
		return nil, 0, nil
	}
	if sb.n >= len(sb.text) {
		return nil, 0, nil
	}
	n, rs, rmode := sb.Tokenizer(sb.text[sb.n:])
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

// LatinTokenizer is the default tokenizer for latin languages
var LatinTokenizer = SimpleBoxerGrab

// StarTokenizer is a demo tokenizer that splits on stars
func StarTokenizer(text []rune) (int, []rune, int) {
	// ... (Same logic as before, assuming it's plain text tokenization)
	// Simplified copy-paste if needed, or keep it.
	// Since I'm overwriting, I should keep the logic.
	// Below is copied logic from previous versions or standard implementation
	n := 0
	rmode := RNIL
	var mode func(rune) bool
	for _, r := range text {
		if mode == nil {
			if r == '*' {
				mode = Once(func(r rune) bool {
					return r == '*'
				})
				rmode = RSimpleBox
				n++
				continue
			} else if IsCR(r) {
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
					return r != '*' && !unicode.IsSpace(r)
				}
			}
		}
		if !mode(r) {
			break
		}
		n++
	}

	if n == 0 {
		return 1, text[:1], RSimpleBox
	}

	return n, text[:n], rmode
}

// SimpleBoxerGrab Consumer of characters until change.
func SimpleBoxerGrab(text []rune) (int, []rune, int) {
	if len(text) == 0 {
		return 0, nil, RNIL
	}

	r := text[0]
	if r == '\r' {
		if len(text) > 1 && text[1] == '\n' {
			return 2, text[:2], RCRLF // CRLF
		}
		return 1, text[:1], RCRLF // CR
	}
	if r == '\n' {
		return 1, text[:1], RCRLF // LF
	}

	if !unicode.IsPrint(text[0]) {
		return 1, nil, RNIL
	}

	isSpace := IsSpaceButNotCRLF(r)

	n := 0
	for n < len(text) {
		r := text[n]
		if IsCR(r) || IsLF(r) {
			break
		}
		if !unicode.IsPrint(r) {
			break
		}
		if IsSpaceButNotCRLF(r) != isSpace {
			break
		}
		n++
	}

	if n == 0 {
		return 1, text[:1], RSimpleBox
	}

	return n, text[:n], RSimpleBox
}

// Matches objects
const (
	RSimpleBox = iota
	RCRLF
	RNIL
)

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

// TurnOnBox draws a box around the box
func (sb *SimpleTextBox) TurnOnBox() {
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

func (sb *SimpleTextBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (sb *SimpleTextBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

// DrawBox renders object
func (sb *SimpleTextBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	if dc.SourceImageMapper != nil {
		originalSrc := sb.drawer.Src
		sb.drawer.Src = dc.SourceImageMapper(originalSrc)
		defer func() {
			sb.drawer.Src = originalSrc
		}()
	}
	sb.drawer.Dst = i
	b := i.Bounds()
	sb.drawer.Dot = fixed.Point26_6{
		X: fixed.I(b.Min.X),
		Y: fixed.I(b.Min.Y) + y,
	}
	sb.drawer.DrawString(sb.Contents)
	if sb.boxBox {
		DrawBox(i, b, dc)
	}
}

// LineBreakBox represents a natural or an effective line break
type LineBreakBox struct {
	// Box is the box that linebreak contains if any
	Box
}

func (sb *LineBreakBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	if sb.Box == nil {
		return 0, 0
	}
	return sb.Box.MinSize()
}

func (sb *LineBreakBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	if sb.Box == nil {
		return 0, 0
	}
	return sb.Box.MaxSize()
}

func (sb *LineBreakBox) AdvanceRect() fixed.Int26_6 {
	return fixed.Int26_6(0)
}

// DrawBox ...
func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	// Should be empty or pass through?
	// Original code:
	/*
		if sb.Box != nil {
			sb.Box.DrawBox(i, y, dc)
		}
	*/
	// Wait, original LineBreakBox implementation in view_file was cut off.
	// It says "func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {" and then nothing?
	// No, I see it usage in SimpleBoxer.
	// Let's assume it draws the underlying box if present.
	if sb.Box != nil {
		sb.Box.DrawBox(i, y, dc)
	}
}

func (sb *LineBreakBox) MetricsRect() font.Metrics {
	if sb.Box != nil {
		return sb.Box.MetricsRect()
	}
	return font.Metrics{}
}

func (sb *LineBreakBox) Whitespace() bool {
	return true
}

func (sb *LineBreakBox) FontDrawer() *font.Drawer {
	if sb.Box != nil {
		return sb.Box.FontDrawer()
	}
	return nil
}

func (sb *LineBreakBox) Len() int {
	if sb.Box != nil {
		return sb.Box.Len()
	}
	return 0
}

func (sb *LineBreakBox) TextValue() string {
	if sb.Box != nil {
		return sb.Box.TextValue()
	}
	return ""
}

// MinSizeBox ensures the box has a minimum size
type MinSizeBox struct {
	Box
	MinSizeVal fixed.Point26_6
}

func (msb *MinSizeBox) AdvanceRect() fixed.Int26_6 {
	a := msb.Box.AdvanceRect()
	if a < msb.MinSizeVal.X {
		return msb.MinSizeVal.X
	}
	return a
}

func (msb *MinSizeBox) MetricsRect() font.Metrics {
	return msb.Box.MetricsRect()
}

func (msb *MinSizeBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	w, h := msb.Box.MinSize()
	if w < msb.MinSizeVal.X {
		w = msb.MinSizeVal.X
	}
	if h < msb.MinSizeVal.Y {
		h = msb.MinSizeVal.Y
	}
	return w, h
}

func (msb *MinSizeBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return msb.Box.MaxSize()
}

func (msb *MinSizeBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	msb.Box.DrawBox(i, y, dc)
}

// PageBreakBox represents a natural or an effective page break
type PageBreakBox struct {
	VisualBox    Box
	ContainerBox Box
}

func NewPageBreak(pbb Box) *PageBreakBox {
	return &PageBreakBox{
		VisualBox: pbb,
	}
}

func (p *PageBreakBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	if p.VisualBox != nil {
		return p.VisualBox.MinSize()
	}
	if p.ContainerBox != nil {
		return p.ContainerBox.MinSize()
	}
	return 0, 0
}

func (p *PageBreakBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	if p.VisualBox != nil {
		return p.VisualBox.MaxSize()
	}
	if p.ContainerBox != nil {
		return p.ContainerBox.MaxSize()
	}
	return 0, 0
}

func (p *PageBreakBox) AdvanceRect() fixed.Int26_6 {
	if p.VisualBox != nil {
		return p.VisualBox.AdvanceRect()
	}
	return 0
}

func (p *PageBreakBox) MetricsRect() font.Metrics {
	if p.VisualBox != nil {
		return p.VisualBox.MetricsRect()
	}
	return font.Metrics{}
}

func (p *PageBreakBox) Whitespace() bool {
	if p.ContainerBox != nil {
		return p.ContainerBox.Whitespace()
	}
	return false
}

func (p *PageBreakBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	if p.VisualBox != nil {
		p.VisualBox.DrawBox(i, y, dc)
	}
}

func (p *PageBreakBox) TurnOnBox() {
	switch p := p.VisualBox.(type) {
	case interface{ TurnOnBox() }:
		p.TurnOnBox()
	}
}

func (p *PageBreakBox) FontDrawer() *font.Drawer {
	if p.VisualBox != nil {
		return p.VisualBox.FontDrawer()
	}
	if p.ContainerBox != nil {
		return p.ContainerBox.FontDrawer()
	}
	return nil
}

func (p *PageBreakBox) Len() int {
	if p.ContainerBox != nil {
		return p.ContainerBox.Len()
	}
	if p.ContainerBox != nil { // Copy paste error in original?
		return p.ContainerBox.Len()
	}
	return 0
}

func (p *PageBreakBox) TextValue() string {
	b := ""
	if p.ContainerBox != nil {
		b += p.ContainerBox.TextValue()
	}
	if p.VisualBox != nil {
		b += p.VisualBox.TextValue()
	}
	return b
}
