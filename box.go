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

func IsCR(r rune) bool {
	return r == '\r'
}

func IsLF(r rune) bool {
	return r == '\n'
}

func (SimpleBoxer) BoxNextWord(fce font.Face, color image.Image, text []rune) (Box, int, error) {
	n := 0
	rs := make([]rune, 0, len(text))
	const (
		RSimpleBox = iota
		RCRLF
	)
	rmode := RSimpleBox
	var mode func(rune) bool
	for _, r := range text {
		if mode == nil {
			if !unicode.IsPrint(r) {
				continue
			}
			if IsCR(r) {
				mode = Once(IsLF)
				n++
				continue
			} else if IsCR(r) {
				rmode = RCRLF
				n++
				break
			} else if IsSpaceButNotCRLF(r) {
				mode = IsSpaceButNotCRLF
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
	switch rmode {
	case RCRLF:
		return &LineBreakBox{}, n, nil
	default:
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
}

func IsSpaceButNotCRLF(r rune) bool {
	return unicode.IsSpace(r) && r != '\n' && r != '\r'
}

func Once(f func(r rune) bool) func(rune) bool {
	c := 0
	return func(r rune) bool {
		c++
		return c == 1 && f(r)
	}
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

type LineBreakBox struct{}

func (sb *LineBreakBox) Image() image.Image {
	return image.NewRGBA(sb.ImageRect())
}

func (sb *LineBreakBox) ImageRect() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{},
	}
}
