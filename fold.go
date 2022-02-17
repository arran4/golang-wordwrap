package wordwrap

import (
	"fmt"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"log"
	"reflect"
)

type Line interface {
	Size() image.Rectangle
	DrawLine(i Image) error
}

type Folder func(b Boxer, pos int, feed []rune) (Line, int, error)

type SimpleLine struct {
	Boxes      []Box
	size       fixed.Rectangle26_6
	fullAscent fixed.Int26_6
}

func (sl *SimpleLine) DrawLine(i Image) error {
	bounds := i.Bounds()
	pmin := bounds.Min
	pmax := bounds.Min
	pmax.Y = bounds.Max.Y
	for _, b := range sl.Boxes {
		switch b := b.(type) {
		case *SimpleBox:
			ir := b.AdvanceRect().Ceil()
			pmax.X += ir
			subImage := i.SubImage(image.Rectangle{
				Min: pmin,
				Max: pmax,
			}).(*image.RGBA)
			b.DrawBox(subImage, sl.fullAscent)
			pmin.X += ir
		}
	}
	return nil
}

func SimpleFolder(boxer Boxer, fce font.Face, feed []rune, container image.Rectangle) (Line, int, error) {
	if len(feed) == 0 {
		return nil, 0, nil
	}
	n := 0
	r := &SimpleLine{
		Boxes: []Box{},
		size:  fixed.R(0, 0, 0, 0),
	}
	done := false
	for !done {
		b, i, err := boxer(fce, image.NewUniform(colornames.Black), feed[n:])
		if err != nil {
			log.Panicf("Error with boxing text: %s", err)
		}
		if b == nil {
			break
		}
		switch b.(type) {
		case *SimpleBox:
			m := b.MetricsRect()
			a := b.AdvanceRect()
			irdx := a.Ceil()
			szdx := (r.size.Max.X - r.size.Min.X).Ceil()
			if irdx+szdx >= container.Dx() {
				if b.Whitespace() {
					n += i
					b = &LineBreakBox{}
				}
				done = true
				break
			}
			r.size.Max.X += a
			ac := -m.Ascent
			if ac < r.size.Min.Y {
				r.size.Min.Y = ac
			}
			dc := m.Descent
			if dc > r.size.Max.Y {
				r.size.Max.Y = dc
			}
			fullAscent := m.Height - m.Descent
			if r.fullAscent < fullAscent {
				r.fullAscent = fullAscent
			}
			log.Printf("%d vs %d", r.fullAscent.Ceil(), -ac)
			n += i
			r.Boxes = append(r.Boxes, b)
		case *LineBreakBox:
			n += i
			done = true
		default:
			return nil, 0, fmt.Errorf("unknown box: %s", reflect.TypeOf(b))
		}
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
