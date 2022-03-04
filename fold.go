package wordwrap

import (
	"bytes"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"reflect"
)

// Line refers to a literal line of text
type Line interface {
	// Size the line consumes
	Size() image.Rectangle
	// DrawLine draws the line
	DrawLine(i Image, options ...DrawOption) error
	// Boxes are the lines contents
	Boxes() []Box
	// TextValue extracts the text value
	TextValue() string
	// YValue where the baseline is
	YValue() int
	// PopSpaceFor will push box at the end, if there isn't enough width, it will make width space.
	PopSpaceFor(sf *SimpleFolder, r image.Rectangle, box Box) error
}

// Folder is the literal line sizer & producer function
type Folder interface {
	// Next line
	Next(yspace int) (Line, error)
}

// SimpleLine is a simple implementation to prevent name space names later. Represents a line
type SimpleLine struct {
	boxes      []Box
	size       fixed.Rectangle26_6
	yoffset    fixed.Int26_6
	boxLine    bool
	fontDrawer *font.Drawer
}

// PopSpaceFor will push box at the end, if there isn't enough width, it will make width space.
func (l *SimpleLine) PopSpaceFor(sf *SimpleFolder, r image.Rectangle, box Box) error {
	ar := box.AdvanceRect()
	lastWs := false
	for r.Dx() < (l.size.Max.X - l.size.Min.X + ar).Ceil() {
		b := l.Pop()
		if b == nil {
			return fmt.Errorf("no more boxes")
		}
		sf.boxer.Unshift(b)
		lastWs = b.Whitespace()
	}
	switch box := box.(type) {
	case *PageBreakBox:
		if lastWs {
			box.ContainerBox = sf.boxer.Shift()
		}
	}
	l.Push(box, ar)
	return nil
}

// Push a box onto the end, and also copy values in appropriately
func (l *SimpleLine) Push(b Box, a fixed.Int26_6) {
	m := b.MetricsRect()
	l.size.Max.X += a
	ac := -m.Ascent
	if ac < l.size.Min.Y {
		l.size.Min.Y = ac
	}
	dc := m.Descent
	if dc > l.size.Max.Y {
		l.size.Max.Y = dc
	}
	yoffset := m.Ascent
	if l.yoffset < yoffset {
		l.yoffset = yoffset
	}
	l.boxes = append(l.boxes, b)
}

// Pop a box off of the end of a line. Ignores all height components that will require a recalculation, drops PageBreak
func (l *SimpleLine) Pop() Box {
	if len(l.boxes) == 0 {
		return nil
	}
	b := l.boxes[len(l.boxes)-1]
	l.boxes = l.boxes[:len(l.boxes)-1]
	a := b.AdvanceRect()
	l.size.Max.X -= a
	for {
		switch box := b.(type) {
		case *LineBreakBox:
			b = box.Box
		default:
			return box
		}
	}
}

var _ Line = (*SimpleLine)(nil)

// Boxes are the lines contents
func (l *SimpleLine) Boxes() []Box {
	return l.boxes
}

// YValue where the baseline is
func (l *SimpleLine) YValue() int {
	return l.yoffset.Ceil()
}

// TextValue extracts the text value of the line
func (l *SimpleLine) TextValue() string {
	b := bytes.NewBuffer(nil)
	for _, e := range l.boxes {
		b.WriteString(e.TextValue())
	}
	return b.String()
}

// turnOnBox turns on drawing a box around the used portion of the line
func (l *SimpleLine) turnOnBox() {
	l.boxLine = true
}

// DrawLine renders image to image, you can control the location by using the SubImage function.
func (l *SimpleLine) DrawLine(i Image, options ...DrawOption) error {
	bounds := i.Bounds()
	r := image.Rectangle{
		Min: bounds.Min,
		Max: bounds.Min,
	}
	r.Max.Y = bounds.Max.Y
	var fi = fixed.I(r.Min.X)
	for _, b := range l.boxes {
		fi += b.AdvanceRect()
		r.Max.X = fi.Round()
		subImage := i.SubImage(r).(Image)
		b.DrawBox(subImage, l.yoffset, options...)
		r.Min.X = r.Max.X
	}
	if l.boxLine {
		DrawBox(i, bounds, options...)
	}
	return nil
}

// OverflowMode Ways of describing overflow
type OverflowMode int

const (
	// StrictBorders default overflow mode. Do not allow
	StrictBorders OverflowMode = iota
	// DescentOverflow Allow some decent overflow. Characters such as yjqp will overflow
	DescentOverflow
	// FullOverflowDuplicate Will allow the full line to overflow, and duplicate line next run
	FullOverflowDuplicate
)

// SimpleFolder is a simple Folder
type SimpleFolder struct {
	boxer          Boxer
	container      image.Rectangle
	lineOptions    []func(Line)
	lastFontDrawer *font.Drawer
	yOverflow      OverflowMode
	// Last object on the last line before a page break if it isn't the last page
	pageBreakBox Box
}

// NewSimpleFolder constructs a SimpleFolder applies options provided.
func NewSimpleFolder(boxer Boxer, container image.Rectangle, lastFontDrawer *font.Drawer, options ...FolderOption) *SimpleFolder {
	r := &SimpleFolder{
		boxer:          boxer,
		container:      container,
		lastFontDrawer: lastFontDrawer,
	}
	for _, option := range options {
		option.ApplyFoldConfig(r)
	}
	return r
}

// Next generates the next life if space
func (sf *SimpleFolder) Next(yspace int) (Line, error) {
	r := &SimpleLine{
		boxes:      []Box{},
		size:       fixed.R(0, 0, 0, 0),
		fontDrawer: sf.lastFontDrawer,
	}
	for {
		b, i, err := sf.boxer.Next()
		if err != nil {
			return nil, fmt.Errorf("boxing at pos %d: %w", sf.boxer.Pos()-i, err)
		}
		if b == nil {
			break
		}

		if r.Size().Dy() < b.MetricsRect().Height.Ceil() {
			rollbackLine := false
			if sf.pageBreakBox != nil && yspace < sf.pageBreakBox.MetricsRect().Height.Ceil() {
				rollbackLine = true
			}
			switch sf.yOverflow {
			case StrictBorders:
				if b.MetricsRect().Height.Ceil() > yspace {
					rollbackLine = true
				}
			}
			if rollbackLine {
				sf.boxer.Push(r.boxes...)
				sf.boxer.Push(b)
				return nil, nil
			}
		}
		done, err := sf.fitAddBox(i, b, r)
		if err != nil {
			return r, err
		}
		if done {
			break
		}
	}
	if len(r.boxes) == 0 {
		return nil, nil
	}
	for _, option := range sf.lineOptions {
		option(r)
	}
	return r, nil
}

// fitAddBox fits if the box and if it does fit adds it. returns new array offset, a bool if it
func (sf *SimpleFolder) fitAddBox(i int, b Box, l *SimpleLine) (bool, error) {
	done := false
	fontDrawer := b.FontDrawer()
	if fontDrawer != nil {
		sf.lastFontDrawer = fontDrawer
	}
	a := b.AdvanceRect()
	switch b.(type) {
	case *SimpleTextBox:
		irdx := a.Ceil()
		szdx := (l.size.Max.X - l.size.Min.X).Ceil()
		cdx := sf.container.Dx()
		if irdx+szdx >= cdx {
			if b.Whitespace() {
				b = &LineBreakBox{
					Box: b,
				}
				l.boxes = append(l.boxes, b)
			} else {
				sf.boxer.Push(b)
			}
			done = true
			return done, nil
		}
	case *LineBreakBox:
		done = true
	default:
		return true, fmt.Errorf("unknown box at pos %d: %s", sf.boxer.Pos()-i, reflect.TypeOf(b))
	}
	l.Push(b, a)
	return done, nil
}

// Size is the size consumed of the line
func (l *SimpleLine) Size() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: (l.size.Max.X - l.size.Min.X).Ceil(),
			Y: (l.size.Max.Y - l.size.Min.Y).Ceil(),
		},
	}
}

// Size is the size consumed of the line
func (sf *SimpleFolder) setPageBreakBox(b Box) {
	sf.pageBreakBox = b
}
