package wordwrap

import (
	"fmt"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"image"
	"log"
	"reflect"
)

type Line interface {
	Size() image.Rectangle
}

type Liner func(b Boxer, pos int, feed []rune) (Line, int, error)

type SimpleLine struct {
	Boxes []Box
	size  image.Rectangle
}

func SimpleLiner(boxer Boxer, fce font.Face, feed []rune, container image.Rectangle) (Line, int, error) {
	if len(feed) == 0 {
		return nil, 0, nil
	}
	n := 0
	r := &SimpleLine{
		Boxes: []Box{},
		size:  image.Rect(0, 0, 0, 0),
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
			ir := b.ImageRect()
			irdx := ir.Dx()
			if irdx+r.size.Dx() >= container.Dx() {
				if b.Whitespace() {
					n += i
					b = &LineBreakBox{}
				}
				done = true
				break
			}
			r.size.Max.X += irdx
			if ir.Min.Y > r.size.Min.Y {
				r.size.Min.Y = ir.Min.Y
			}
			if ir.Max.Y > r.size.Max.Y {
				r.size.Max.Y = ir.Max.Y
			}
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
	return sl.size
}
