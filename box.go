package wordwrap

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
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
// SimpleBoxer simple tokenizer basically determines if something unicode.IsSpace or is a new line, or is text and tells
// the calling Folder that. Putting the elements in the correct Box.
type SimpleBoxer struct {
	postBoxOptions []func(Box)
	contents       []*Content
	contentIndex   int
	n              int
	fontDrawer     *font.Drawer
	Tokenizer      Tokenizer
	cacheQueue     []Box
}

// Ensures that SimpleBoxer fits model
var _ Boxer = (*SimpleBoxer)(nil)

// RichBoxer alias for backward compatibility
type RichBoxer = SimpleBoxer

// NewSimpleBoxer simple tokenizer basically determines if something unicode.IsSpace or is a new line, or is text and tells
// the calling Folder that. Putting the elements in the correct Box.
func NewSimpleBoxer(contents []*Content, drawer *font.Drawer, options ...BoxerOption) *SimpleBoxer {
	sb := &SimpleBoxer{
		contents:   contents,
		n:          0,
		fontDrawer: drawer,
		Tokenizer:  LatinTokenizer,
	}
	for _, option := range options {
		option.ApplyBoxConfig(sb)
	}
	return sb
}

// NewRichBoxer creates a new boxer using a variety of arguments to create the contents and options
func NewRichBoxer(args ...interface{}) *SimpleBoxer {
	contents, drawer, _, boxerOptions, _, tokenizer := ProcessRichArgs(args...)

	sb := &SimpleBoxer{
		contents:   contents,
		n:          0,
		fontDrawer: drawer,
		Tokenizer:  LatinTokenizer,
	}
	if tokenizer != nil {
		sb.Tokenizer = tokenizer
	}
	for _, option := range boxerOptions {
		option.ApplyBoxConfig(sb)
	}
	return sb
}

// Reset restarts the tokenization
func (sb *SimpleBoxer) Reset() {
	sb.n = 0
	sb.contentIndex = 0
	sb.cacheQueue = nil
}

// Pos current parser position.
func (sb *SimpleBoxer) Pos() int {
	pos := 0
	for i := 0; i < sb.contentIndex; i++ {
		pos += len([]rune(sb.contents[i].text))
	}
	pos += sb.n

	for _, e := range sb.cacheQueue {
		pos -= e.Len()
	}
	return pos
}

// HasNext unprocessed bytes exist
func (sb *SimpleBoxer) HasNext() bool {
	if len(sb.cacheQueue) > 0 {
		return true
	}
	for i := sb.contentIndex; i < len(sb.contents); i++ {
		text := []rune(sb.contents[i].text)
		start := 0
		if i == sb.contentIndex {
			start = sb.n
		}
		if len(sb.contents[i].children) > 0 && start == 0 {
			return true
		}
		if sb.contents[i].image != nil && start == 0 {
			return true
		}
		if start < len(text) {
			return true
		}
	}
	return false
}

// SetFontDrawer Changes the default font
func (sb *SimpleBoxer) SetFontDrawer(face *font.Drawer) {
	sb.fontDrawer = face
}

// Back goes back i spaces (ie unreads)
func (sb *SimpleBoxer) Back(i int) {
	sb.n -= i
	for sb.n < 0 {
		sb.contentIndex--
		if sb.contentIndex < 0 {
			sb.contentIndex = 0
			sb.n = 0
			break
		}
		sb.n += len([]rune(sb.contents[sb.contentIndex].text))
	}
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
		// Height/CapHeight/XHeight? Use max or copy first?
		// Usually max ascent/descent is what matters for line box.
	}
	return m
}

func (rb *RowBox) Whitespace() bool {
	// If any box is NOT whitespace, then the row is NOT whitespace.
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
		// Draw each box, using a sub-image for proper relative positioning.

		w := b.AdvanceRect()
		adv := w.Ceil()
		r := i.Bounds()
		// Current slot: [r.Min.X + x, r.Min.Y, r.Min.X + x + adv, r.Max.Y]
		// Clip to bounds
		minX := r.Min.X + x.Ceil()
		maxX := minX + adv
		if maxX > r.Max.X {
			maxX = r.Max.X
		}
		// Y is usually full height?
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
	for {
		if sb.contentIndex >= len(sb.contents) {
			return nil, 0, nil
		}
		currentContent := sb.contents[sb.contentIndex]

		if len(currentContent.children) > 0 {
			sb.contentIndex++
			sb.n = 0

			// Process children into boxes
			subBoxer := NewSimpleBoxer(currentContent.children, sb.fontDrawer)
			var boxes []Box
			for subBoxer.HasNext() {
				b, _, err := subBoxer.Next()
				if err != nil {
					return nil, 0, err
				}
				if b != nil {
					boxes = append(boxes, b)
				}
			}

			var b Box = &RowBox{Boxes: boxes}

			// Apply decorators
			if currentContent.style != nil {
				if currentContent.style.Alignment != AlignBaseline {
					b = &AlignedBox{
						Box:       b,
						Alignment: currentContent.style.Alignment,
					}
				}
				if !currentContent.style.Padding.Empty() || !currentContent.style.Margin.Empty() {
					bg := currentContent.style.BackgroundColor
					b = NewDecorationBox(b, currentContent.style.Padding, currentContent.style.Margin, bg, currentContent.style.BgPositioning)
				}
			}
			for i := len(currentContent.decorators) - 1; i >= 0; i-- {
				b = currentContent.decorators[i](b)
			}
			if currentContent.id != nil {
				b = &IDBox{
					Box: b,
					id:  currentContent.id,
				}
			}
			return b, 0, nil
		}

		if sb.n == 0 && currentContent.image != nil {
			sb.contentIndex++
			sb.n = 0
			var b Box
			b = &ImageBox{
				I: currentContent.image,
			}
			if currentContent.imageScale != 0 {
				b = &ImageBox{
					I:     currentContent.image,
					Scale: currentContent.imageScale,
				}
			}
			if currentContent.style != nil {
				if currentContent.style.BackgroundColor != nil {
					b = &BackgroundBox{
						Box:        b,
						Background: currentContent.style.BackgroundColor,
					}
				}
				// Effects
				if len(currentContent.style.Effects) > 0 {
					b = &EffectBox{
						Box:     b,
						Effects: currentContent.style.Effects,
					}
				}
				if currentContent.style.Alignment != AlignBaseline {
					b = &AlignedBox{
						Box:       b,
						Alignment: currentContent.style.Alignment,
					}
				}
				if !currentContent.style.Padding.Empty() || !currentContent.style.Margin.Empty() {
					bg := currentContent.style.BackgroundColor
					b = NewDecorationBox(b, currentContent.style.Padding, currentContent.style.Margin, bg, currentContent.style.BgPositioning)
				}
				for i := len(currentContent.decorators) - 1; i >= 0; i-- {
					b = currentContent.decorators[i](b)
				}
			}
			if currentContent.id != nil {
				b = &IDBox{
					Box: b,
					id:  currentContent.id,
				}
			}
			for _, option := range sb.postBoxOptions {
				option(b)
			}
			return b, 1, nil
		}
		text := []rune(currentContent.text)
		if sb.n >= len(text) {
			sb.contentIndex++
			sb.n = 0
			continue
		}
		n, rs, rmode := sb.Tokenizer(text[sb.n:])
		sb.n += n
		var b Box
		drawer := sb.fontDrawer
		if currentContent.style != nil {
			if currentContent.style.font != nil {
				drawer = &font.Drawer{
					Src:  sb.fontDrawer.Src,
					Face: currentContent.style.font,
				}
			}
			if currentContent.style.FontDrawerSrc != nil {
				if drawer == sb.fontDrawer {
					drawer = &font.Drawer{
						Src:  sb.fontDrawer.Src,
						Face: sb.fontDrawer.Face,
					}
				}
				drawer.Src = currentContent.style.FontDrawerSrc
			}
		}

		switch rmode {
		case RNIL:
			if sb.n >= len(text) {
				sb.contentIndex++
				sb.n = 0
				if sb.contentIndex >= len(sb.contents) {
					return nil, 0, nil
				}
				continue
			}
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
		if currentContent.style != nil {
			if currentContent.style.BackgroundColor != nil {
				b = &BackgroundBox{
					Box:           b,
					Background:    currentContent.style.BackgroundColor,
					BgPositioning: currentContent.style.BgPositioning,
				}
			}
			if len(currentContent.style.Effects) > 0 {
				b = &EffectBox{
					Box:     b,
					Effects: currentContent.style.Effects,
				}
			}
			if currentContent.style.Alignment != AlignBaseline {
				b = &AlignedBox{
					Box:       b,
					Alignment: currentContent.style.Alignment,
				}
			}
			if !currentContent.style.Padding.Empty() || !currentContent.style.Margin.Empty() {
				bg := currentContent.style.BackgroundColor
				b = NewDecorationBox(b, currentContent.style.Padding, currentContent.style.Margin, bg, currentContent.style.BgPositioning)
			}
			for i := len(currentContent.decorators) - 1; i >= 0; i-- {
				b = currentContent.decorators[i](b)
			}
		}
		if currentContent.id != nil {
			b = &IDBox{
				Box: b,
				id:  currentContent.id,
			}
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
}

// LatinTokenizer is the default tokenizer for latin languages
var LatinTokenizer = SimpleBoxerGrab

// StarTokenizer is a demo tokenizer that splits on stars
func StarTokenizer(text []rune) (int, []rune, int) {
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

	if n == 0 { // Should not happen given the checks above, but as a safeguard.
		return 1, text[:1], RSimpleBox
	}

	return n, text[:n], rmode
}

// SimpleBoxerGrab Consumer of characters until change. Could be made to conform to strings.Scanner
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
		// Consume a single non-printable character and signal to ignore it.
		// This prevents infinite loops on non-printable characters.
		return 1, nil, RNIL
	}

	isSpace := IsSpaceButNotCRLF(r)

	n := 0
	for n < len(text) {
		r := text[n]
		if IsCR(r) || IsLF(r) {
			break
		}
		// Also stop at non-printable characters.
		if !unicode.IsPrint(r) {
			break
		}
		if IsSpaceButNotCRLF(r) != isSpace {
			break
		}
		n++
	}

	if n == 0 { // Should not happen given the checks above, but as a safeguard.
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

// DrawBox renders object

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
	// TODO: Handle MinHeight?
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
	// Draw inner box.
	// Default alignment: Left.
	msb.Box.DrawBox(i, y, dc)
}
func (sb *LineBreakBox) AdvanceRect() fixed.Int26_6 {
	return fixed.Int26_6(0)
}

// PageBreakBox represents a natural or an effective page break
type PageBreakBox struct {
	// VisualBox is the box to render and use
	VisualBox Box
	// ContainerBox is the box that linebreak contains if any
	ContainerBox Box
}

// NewPageBreak basic constructor for a page break.
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

// AdvanceRect width of text
func (p *PageBreakBox) AdvanceRect() fixed.Int26_6 {
	if p.VisualBox != nil {
		return p.VisualBox.AdvanceRect()
	}
	return 0
}

// MetricsRect all other font details of text
func (p *PageBreakBox) MetricsRect() font.Metrics {
	if p.VisualBox != nil {
		return p.VisualBox.MetricsRect()
	}
	return font.Metrics{}
}

// Whitespace if contains a white space or not
func (p *PageBreakBox) Whitespace() bool {
	if p.ContainerBox != nil {
		return p.ContainerBox.Whitespace()
	}
	return false
}

// DrawBox renders object
func (p *PageBreakBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	if p.VisualBox != nil {
		p.VisualBox.DrawBox(i, y, dc)
	}
}

// turnOnBox draws a box around the box
func (p *PageBreakBox) turnOnBox() {
	switch p := p.VisualBox.(type) {
	case interface{ turnOnBox() }:
		p.turnOnBox()
	}
}

// FontDrawer font used
func (p *PageBreakBox) FontDrawer() *font.Drawer {
	if p.VisualBox != nil {
		return p.VisualBox.FontDrawer()
	}
	if p.ContainerBox != nil {
		return p.ContainerBox.FontDrawer()
	}
	return nil
}

// Len the length of the buffer represented by the box
func (p *PageBreakBox) Len() int {
	if p.ContainerBox != nil {
		return p.ContainerBox.Len()
	}
	if p.ContainerBox != nil {
		return p.ContainerBox.Len()
	}
	return 0
}

// TextValue returns the text suppressed by the line break (probably a white space including a \r\n)
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

// IDBox wrapper for box with ID
type IDBox struct {
	Box
	id interface{}
}

// ID returns the ID
func (ib *IDBox) ID() interface{} {
	return ib.id
}

// DrawBox renders object

func (b *IDBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MinSize()
}

func (b *IDBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MaxSize()
}

func (ib *IDBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	ib.Box.DrawBox(i, y, dc)
}

// Interface enforcement
var _ Box = (*IDBox)(nil)

// ImageBox is a box that contains an image
type ImageBox struct {
	I          image.Image
	Scale      float64
	M          font.Metrics
	metricCalc imageBoxOptionMetricCalcFunc
	fontDrawer *font.Drawer
	boxBox     bool
}

// Interface enforcement
var _ Box = (*ImageBox)(nil)

// AdvanceRect width of text
func (ib *ImageBox) AdvanceRect() fixed.Int26_6 {
	scale := ib.Scale
	if scale == 0 {
		scale = 1
	}
	return fixed.I(int(float64(ib.I.Bounds().Dx()) * scale))
}

// MetricsRect all other font details of text
func (ib *ImageBox) MetricsRect() font.Metrics {
	return ib.M
}

// Whitespace if this is a white space or not
func (ib *ImageBox) Whitespace() bool {
	return false
}

// DrawBox renders object
func (ib *ImageBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	bounds := i.Bounds()
	srci := ib.I
	if dc.SourceImageMapper != nil {
		originalSrc := srci
		srci = dc.SourceImageMapper(originalSrc)
		defer func() {
			srci = originalSrc
		}()
	}
	draw.Draw(i, bounds.Add(image.Pt(0, (y - ib.M.Ascent).Ceil())), srci, srci.Bounds().Min, draw.Over)
	if ib.boxBox {
		DrawBox(i, bounds, dc)
	}
}

// turnOnBox draws a box around the box
func (ib *ImageBox) turnOnBox() {
	ib.boxBox = true
}

// FontDrawer font used
func (ib *ImageBox) FontDrawer() *font.Drawer {
	return ib.fontDrawer
}

// Len the length of the buffer represented by the box

func (ib *ImageBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (ib *ImageBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (ib *ImageBox) Len() int {
	return 0
}

// TextValue returns the text suppressed by the line break (probably a white space including a \r\n)
func (ib *ImageBox) TextValue() string {
	return ""
}

// CalculateMetrics calculate dimension and positioning
func (ib *ImageBox) CalculateMetrics() {
	if ib.metricCalc == nil {
		ib.M = ImageBoxMetricAboveTheLine(ib)
	} else {
		ib.M = ib.metricCalc(ib)
	}
}

// NewImageBox constructs a new ImageBox
func NewImageBox(i image.Image, options ...ImageBoxOption) *ImageBox {
	ib := &ImageBox{
		I: i,
	}
	for _, o := range options {
		o.applyImageBoxOption(ib)
	}
	ib.CalculateMetrics()
	return ib
}

// BackgroundBox is a box that has a background
type BackgroundBox struct {
	Box
	Background    image.Image
	BgPositioning BackgroundPositioning
	boxBox        bool
}

// DrawBox renders object

func (b *BackgroundBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MinSize()
}

func (b *BackgroundBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MaxSize()
}

func (bb *BackgroundBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	bounds := i.Bounds()
	srcPoint := image.Point{}
	switch bb.BgPositioning {
	case BgPositioningPassThrough:
		srcPoint = bounds.Min
	case BgPositioningZeroed:
		srcPoint = image.Point{}
	case BgPositioningSection5Zeroed:
		// Since BackgroundBox usually wraps the content directly with no padding/margin difference in this context,
		// Section5Zeroed (content relative) is equivalent to Zeroed (box relative) if no margin/pad.
		// Or effectively 0,0 relative to Bounds.Min?
		// Section5Zeroed means "Match section 5 (content) starting position with coordinates 0, 0".
		// Here, content starts at bounds.Min. so 0 = bounds.Min - bounds.Min + srcPoint.
		// srcPoint = 0.
		srcPoint = image.Point{}
	}

	// Draw background
	draw.Draw(i, bounds, bb.Background, srcPoint, draw.Over)
	// Draw content
	bb.Box.DrawBox(i, y, dc)
	if bb.boxBox {
		DrawBox(i, bounds, dc)
	}
}

// turnOnBox draws a box around the box
func (bb *BackgroundBox) turnOnBox() {
	bb.boxBox = true
	// We might want to pass it down?
	if b, ok := bb.Box.(interface{ turnOnBox() }); ok {
		b.turnOnBox()
	}
}

// EffectType defines when the effect is applied
type EffectType int

const (
	EffectPre EffectType = iota
	EffectPost
)

// BoxEffect defines a graphical effect applied to a box
type BoxEffect struct {
	Func func(Image, Box, *DrawConfig)
	Type EffectType
}

// EffectBox wraps a box with effects
type EffectBox struct {
	Box
	Effects []BoxEffect
	boxBox  bool
}

// DrawBox renders object with effects

func (b *EffectBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MinSize()
}

func (b *EffectBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MaxSize()
}

func (eb *EffectBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	for _, e := range eb.Effects {
		if e.Type == EffectPre {
			e.Func(i, eb.Box, dc)
		}
	}
	eb.Box.DrawBox(i, y, dc)
	for _, e := range eb.Effects {
		if e.Type == EffectPost {
			e.Func(i, eb.Box, dc)
		}
	}
	if eb.boxBox {
		DrawBox(i, i.Bounds(), dc)
	}
}

// turnOnBox draws a box around the box
func (eb *EffectBox) turnOnBox() {
	eb.boxBox = true
	// Pass down
	if b, ok := eb.Box.(interface{ turnOnBox() }); ok {
		b.turnOnBox()
	}
}

// Interface enforcement
var _ Box = (*EffectBox)(nil)

// AlignedBox wraps a box with alignment information
type AlignedBox struct {
	Box
	Alignment BaselineAlignment
}

// DrawBox renders object (passes through, logic is in Line)

func (b *AlignedBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MinSize()
}

func (b *AlignedBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return b.Box.MaxSize()
}

func (ab *AlignedBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	ab.Box.DrawBox(i, y, dc)
}

// turnOnBox draws a box around the box
func (ab *AlignedBox) turnOnBox() {
	if b, ok := ab.Box.(interface{ turnOnBox() }); ok {
		b.turnOnBox()
	}
}

// Interface enforcement
var _ Box = (*AlignedBox)(nil)

// FillLineMode determines how the box consumes the line
type FillLineMode int

const (
	// FillRestOfLine consumes the rest of the current line
	FillRestOfLine FillLineMode = iota
	// FillEntireLine consumes the entire line (starts on new line, ends line)
	FillEntireLine
)

// FillLineBox is a box that consumes the line
type FillLineBox struct {
	Box
	Mode FillLineMode
}

// NewFillLineBox constructs a new FillLineBox
func NewFillLineBox(b Box, mode FillLineMode) *FillLineBox {
	return &FillLineBox{
		Box:  b,
		Mode: mode,
	}
}

// DrawBox delegates to the inner box.
func (flb *FillLineBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	flb.Box.DrawBox(i, y, dc)
}

// Ensure interface compliance
var _ Box = (*FillLineBox)(nil)
