package wordwrap

import (
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
	DrawLine(i Image) error
	// Boxes are the lines contents
	Boxes() []Box
}

// Folder is the literal line sizer & producer function
type Folder interface {
	// Next line
	Next() (Line, error)
}

// SimpleLine is a simple implementation to prevent name space names later. Represents a line
type SimpleLine struct {
	boxes      []Box
	size       fixed.Rectangle26_6
	height     fixed.Int26_6
	boxLine    bool
	fontDrawer *font.Drawer
}

// Boxes are the lines contents
func (sl *SimpleLine) Boxes() []Box {
	return sl.boxes
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
		b.DrawBox(subImage, sl.height)
		r.Min.X = r.Max.X
	}
	if sl.boxLine {
		util.DrawBox(i, bounds)
	}
	return nil
}

// SimpleFolder is a simple Folder
type SimpleFolder struct {
	boxer          Boxer
	container      image.Rectangle
	lineOptions    []func(Line)
	lastFontDrawer *font.Drawer
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
func (sf *SimpleFolder) Next() (Line, error) {
	n := 0
	r := &SimpleLine{
		boxes:      []Box{},
		size:       fixed.R(0, 0, 0, 0),
		fontDrawer: sf.lastFontDrawer,
	}
	for {
		b, i, err := sf.boxer.Next()
		if err != nil {
			return nil, fmt.Errorf("boxing at pos %d: %w", n-i, err)
		}
		if b == nil {
			break
		}
		nn, done, err := sf.fitAddBox(n, i, b, r)
		if err != nil {
			return r, err
		}
		n = nn
		if done {
			break
		}
	}
	if n == 0 {
		return nil, nil
	}
	for _, option := range sf.lineOptions {
		option(r)
	}
	return r, nil
}

// fitAddBox fits if the box and if it does fit adds it. returns new array offset, a bool if it
func (sf *SimpleFolder) fitAddBox(n int, i int, b Box, l *SimpleLine) (int, bool, error) {
	done := false
	n += i
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
		if irdx+szdx >= sf.container.Dx() {
			if b.Whitespace() {
				b = &LineBreakBox{
					fontDrawer: fontDrawer,
				}
				l.boxes = append(l.boxes, b)
			} else {
				sf.boxer.Back(i)
				n -= i
			}
			done = true
			return n, done, nil
		}
		l.size.Max.X += a
	case *LineBreakBox:
		done = true
	default:
		return 0, true, fmt.Errorf("unknown box at pos %d: %s", n-i, reflect.TypeOf(b))
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
	if l.height < height {
		l.height = height
	}
	l.boxes = append(l.boxes, b)
	return n, done, nil
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
