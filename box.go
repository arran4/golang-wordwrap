package wordwrap

import (
	"errors"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"unicode"
)

type Box interface {
	ImageRect() image.Rectangle
	Image() image.Image
}

type SimpleBoxer struct{}

type Boxer interface {
	BoxNextWord(fce font.Face, color image.Image, text []rune) (Box, int, error)
}

func NewSimpleBoxer() Boxer {
	return &SimpleBoxer{}
}

func (SimpleBoxer) BoxNextWord(fce font.Face, color image.Image, text []rune) (Box, int, error) {
	n := 0
	rs := make([]rune, 0, len(text))
	var mode func(rune) bool
	for _, r := range text {
		if mode == nil {
			if !unicode.IsPrint(r) {
				continue
			}
			if unicode.IsSpace(r) {
				mode = unicode.IsSpace
			} else {
				mode = func(r rune) bool {
					return !unicode.IsSpace(r)
				}
			}
		}
		if !mode(r) {
			break
		}
		rs = append(rs, r)
		n++
	}
	t := string(rs)
	drawer := &font.Drawer{
		Src:  color,
		Face: fce,
	}
	if fce == nil {
		return nil, 0, errors.New("font face not provided")
	}
	ttb, _ := drawer.BoundString(t)
	return &SimpleBox{
		drawer:   drawer,
		Contents: t,
		Size:     ttb,
	}, n, nil
}

type SimpleBox struct {
	Contents string
	Size     fixed.Rectangle26_6
	drawer   *font.Drawer
}

func (sb *SimpleBox) Image() image.Image {
	i := image.NewRGBA(sb.ImageRect())
	if sb.drawer == nil {
		return i
	}
	sb.drawer.Dst = i
	sb.drawer.Dot = sb.drawer.Dot.Sub(sb.Size.Min)
	sb.drawer.DrawString(sb.Contents)
	return i
}

func (sb *SimpleBox) ImageRect() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: sb.Size.Min.X.Round(),
			Y: sb.Size.Min.Y.Round(),
		},
		Max: image.Point{
			X: sb.Size.Max.X.Round(),
			Y: sb.Size.Max.Y.Round(),
		},
	}
}
