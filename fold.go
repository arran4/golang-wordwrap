package wordwrap

import (
	"fmt"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"reflect"
)

type Line interface {
	Size() image.Rectangle
	DrawLine(i Image) error
}

type Folder func(b Boxer, pos int, feed []rune, options ...FolderOption) (Line, int, error)

type SimpleLine struct {
	Boxes        []Box
	size         fixed.Rectangle26_6
	height       fixed.Int26_6
	boxLine      bool
	boxerOptions []BoxerOption
}

func (sl *SimpleLine) addBoxConfig(bo BoxerOption) {
	sl.boxerOptions = append(sl.boxerOptions, bo)
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
	for _, b := range sl.Boxes {
		fi += b.AdvanceRect()
		r.Max.X = fi.Round()
		subImage := i.SubImage(r).(*image.RGBA)
		b.DrawBox(subImage, sl.height)
		r.Min.X = r.Max.X
	}
	if sl.boxLine {
		util.DrawBox(i, bounds)
	}
	return nil
}

func SimpleFolder(boxer Boxer, fce font.Face, feed []rune, container image.Rectangle, options ...FolderOption) (Line, int, error) {
	if len(feed) == 0 {
		return nil, 0, nil
	}
	n := 0
	r := &SimpleLine{
		Boxes: []Box{},
		size:  fixed.R(0, 0, 0, 0),
	}
	for _, option := range options {
		option.ApplyFoldConfig(r)
	}
	done := false
	for !done {
		b, i, err := boxer(fce, image.NewUniform(colornames.Black), feed[n:], r.boxerOptions...)
		if err != nil {
			return nil, 0, fmt.Errorf("boxing %d %w", n, err)
		}
		if b == nil {
			break
		}
		m := b.MetricsRect()
		switch b.(type) {
		case *SimpleBox:
			a := b.AdvanceRect()
			irdx := a.Ceil()
			szdx := (r.size.Max.X - r.size.Min.X).Ceil()
			if irdx+szdx >= container.Dx() {
				if b.Whitespace() {
					b = &LineBreakBox{
						fce: fce,
					}
					n += i
					r.Boxes = append(r.Boxes, b)
				}
				done = true
				continue
			}
			r.size.Max.X += a
		case *LineBreakBox:
			done = true
		default:
			return nil, 0, fmt.Errorf("unknown box: %s", reflect.TypeOf(b))
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
		n += i
		r.Boxes = append(r.Boxes, b)
	}
	return r, n, nil
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
