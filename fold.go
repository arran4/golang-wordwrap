package wordwrap

import (
	"bytes"
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
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
	DrawLine(i Image) error
	// Boxes are the lines contents
	Boxes() []Box
	// TextValue extracts the text value
	TextValue() string
	// YValue where the baseline is
	YValue() int
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

var _ Line = (*SimpleLine)(nil)

// Boxes are the lines contents
func (sl *SimpleLine) Boxes() []Box {
	return sl.boxes
}

// YValue where the baseline is
func (sl *SimpleLine) YValue() int {
	return sl.yoffset.Ceil()
}

// TextValue extracts the text value of the line
func (sl *SimpleLine) TextValue() string {
	b := bytes.NewBuffer(nil)
	for _, e := range sl.boxes {
		b.WriteString(e.TextValue())
	}
	return b.String()
}

// turnOnBox turns on drawing a box around the used portion of the line
func (sl *SimpleLine) turnOnBox() {
	sl.boxLine = true
}

// DrawLine renders image to image, you can control the location by using the SubImage function.
func (sl *SimpleLine) DrawLine(i Image) error {
	bounds := i.Bounds()
	r := image.Rectangle{
		Min: bounds.Min,
		Max: bounds.Min,
	}
	r.Max.Y = bounds.Max.Y
	var fi = fixed.I(r.Min.X)
	for _, b := range sl.boxes {
		fi += b.AdvanceRect()
		r.Max.X = fi.Round()
		subImage := i.SubImage(r).(Image)
		b.DrawBox(subImage, sl.yoffset)
		r.Min.X = r.Max.X
	}
	if sl.boxLine {
		util.DrawBox(i, bounds)
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

		sf.pageBreakBox

		if r.Size().Dy() < b.MetricsRect().Height.Ceil() {
			switch sf.yOverflow {
			case StrictBorders:
				if b.MetricsRect().Height.Ceil() > yspace {
					sf.boxer.Push(r.boxes...)
					sf.boxer.Push(b)
					return nil, nil
				}
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
	m := b.MetricsRect()
	fontDrawer := b.FontDrawer()
	if fontDrawer != nil {
		sf.lastFontDrawer = fontDrawer
	} else {
		fontDrawer = sf.lastFontDrawer
	}
	switch b.(type) {
	case *SimpleBox:
		a := b.AdvanceRect()
		irdx := a.Ceil()
		szdx := (l.size.Max.X - l.size.Min.X).Ceil()
		cdx := sf.container.Dx()
		if irdx+szdx >= cdx {
			if b.Whitespace() {
				b = &LineBreakBox{
					fontDrawer: fontDrawer,
					text:       b.TextValue(),
				}
				l.boxes = append(l.boxes, b)
			} else {
				sf.boxer.Push(b)
			}
			done = true
			return done, nil
		}
		l.size.Max.X += a
	case *LineBreakBox:
		done = true
	default:
		return true, fmt.Errorf("unknown box at pos %d: %s", sf.boxer.Pos()-i, reflect.TypeOf(b))
	}
	ac := -m.Ascent
	if ac < l.size.Min.Y {
		l.size.Min.Y = ac
	}
	dc := m.Descent
	if dc > l.size.Max.Y {
		l.size.Max.Y = dc
	}
	height := m.Ascent
	if l.yoffset < height {
		l.yoffset = height
	}
	l.boxes = append(l.boxes, b)
	return done, nil
}

// Size is the size consumed of the line
func (sl *SimpleLine) Size() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: (sl.size.Max.X - sl.size.Min.X).Ceil(),
			Y: (sl.size.Max.Y - sl.size.Min.Y).Ceil(),
		},
	}
}

// Size is the size consumed of the line
func (sf *SimpleFolder) setPageBreakChevron(i image.Image) {
	sf.pageBreakBox = NewImageBox(i)
}
