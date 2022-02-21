package wordwrap

import (
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"reflect"
)

type Line interface {
	Size() image.Rectangle
	DrawLine(i Image) error
	Boxes() []Box
}

type Folder interface {
	Next() (Line, error)
}

type SimpleLine struct {
	boxes        []Box
	size         fixed.Rectangle26_6
	height       fixed.Int26_6
	boxLine      bool
	boxerOptions []BoxerOption
}

func (sl *SimpleLine) addBoxConfig(bo BoxerOption) {
	sl.boxerOptions = append(sl.boxerOptions, bo)
}

func (sl *SimpleLine) Boxes() []Box {
	return sl.boxes
}

func (sl *SimpleLine) turnOnBox() {
	sl.boxLine = true
}

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

type SimpleFolder struct {
	boxer       Boxer
	container   image.Rectangle
	lineOptions []func(Line)
}

func NewSimpleFolder(boxer Boxer, container image.Rectangle, options ...FolderOption) *SimpleFolder {
	r := &SimpleFolder{
		boxer:     boxer,
		container: container,
	}
	for _, option := range options {
		option.ApplyFoldConfig(r)
	}
	return r
}

func (sf *SimpleFolder) Next() (Line, error) {
	n := 0
	r := &SimpleLine{
		boxes: []Box{},
		size:  fixed.R(0, 0, 0, 0),
	}
	var lastFont *font.Drawer
	done := false
	for !done {
		b, i, err := sf.boxer.Next()
		if err != nil {
			return nil, fmt.Errorf("boxing at pos %d: %w", n-i, err)
		}
		if b == nil {
			break
		}
		n += i
		lastFont = b.FontDrawer()
		m := b.MetricsRect()
		switch b.(type) {
		case *SimpleBox:
			a := b.AdvanceRect()
			irdx := a.Ceil()
			szdx := (r.size.Max.X - r.size.Min.X).Ceil()
			if irdx+szdx >= sf.container.Dx() {
				if b.Whitespace() {
					b = &LineBreakBox{
						fontDrawer: lastFont,
					}
					r.boxes = append(r.boxes, b)
				} else {
					sf.boxer.Back(i)
					n -= i
				}
				done = true
				continue
			}
			r.size.Max.X += a
		case *LineBreakBox:
			done = true
		default:
			return nil, fmt.Errorf("unknown box at pos %d: %s", n-i, reflect.TypeOf(b))
		}
		ac := -m.Ascent
		if ac < r.size.Min.Y {
			r.size.Min.Y = ac
		}
		dc := m.Descent
		if dc > r.size.Max.Y {
			r.size.Max.Y = dc
		}
		height := m.Ascent
		if r.height < height {
			r.height = height
		}
		r.boxes = append(r.boxes, b)
	}
	if n == 0 {
		return nil, nil
	}
	for _, option := range sf.lineOptions {
		option(r)
	}
	return r, nil
}

func (sl *SimpleLine) Size() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: (sl.size.Max.X - sl.size.Min.X).Ceil(),
			Y: (sl.size.Max.Y - sl.size.Min.Y).Ceil(),
		},
	}
}
