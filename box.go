package wordwrap

import (
	"errors"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"unicode"
)

type Box interface {
	ImageRect() image.Rectangle
	Image() image.Image
	FontRect() fixed.Rectangle26_6
	Whitespace() bool
}

type Boxer func(fce font.Face, color image.Image, text []rune) (Box, int, error)

func IsCR(r rune) bool {
	return r == '\r'
}

func IsLF(r rune) bool {
	return r == '\n'
}

func SimpleBoxer(fce font.Face, color image.Image, text []rune) (Box, int, error) {
	n, rs, rmode := SimpleBoxerGrab(text)
	switch rmode {
	case RNIL:
		return nil, n, nil
	case RCRLF:
		return &LineBreakBox{}, n, nil
	case RSimpleBox:
		t := string(rs)
		drawer := &font.Drawer{
			Src:  color,
			Face: fce,
		}
		if fce == nil {
			return nil, 0, errors.New("font face not provided")
		}
		ttb, a := drawer.BoundString(t)
		return &SimpleBox{
			drawer:   drawer,
			Contents: t,
			Bounds:   ttb,
			Advance:  a,
		}, n, nil
	default:
		return nil, 0, fmt.Errorf("unknown rmode %d", rmode)
	}
}

const (
	RSimpleBox = iota
	RCRLF
	RNIL
)

func SimpleBoxerGrab(text []rune) (int, []rune, int) {
	n := 0
	rs := make([]rune, 0, len(text))
	rmode := RNIL
	var mode func(rune) bool
	for _, r := range text {
		if mode == nil {
			if IsCR(r) {
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
	return n, rs, rmode
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
	Bounds   fixed.Rectangle26_6
	drawer   *font.Drawer
	Advance  fixed.Int26_6
}

func (sb *SimpleBox) FontRect() fixed.Rectangle26_6 {
	return sb.Bounds
}

func (sb *SimpleBox) Whitespace() bool {
	return sb.Contents == "" || unicode.IsSpace(rune(sb.Contents[0]))
}

func (sb *SimpleBox) Image() image.Image {
	i := image.NewRGBA(sb.ImageRect())
	if sb.drawer == nil {
		return i
	}
	sb.drawer.Dst = i
	sb.drawer.Dot = sb.drawer.Dot.Sub(sb.Bounds.Min)
	sb.drawer.DrawString(sb.Contents)
	return i
}

func (sb *SimpleBox) ImageRect() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: -sb.drawer.Face.Metrics().Ascent.Round(),
		},
		Max: image.Point{
			X: (sb.Advance).Round(),
			Y: sb.drawer.Face.Metrics().Height.Round() - sb.drawer.Face.Metrics().Ascent.Round(),
		},
	}
}

func (sb *SimpleBox) DrawBox(i Image, y fixed.Int26_6) {
	sb.drawer.Dst = i
	b := i.Bounds()
	sb.drawer.Dot = fixed.Point26_6{
		X: fixed.I(b.Min.X),
		Y: fixed.I(b.Min.Y) + y,
	}
	sb.drawer.DrawString(sb.Contents)
}

type LineBreakBox struct{}

func (sb *LineBreakBox) FontRect() fixed.Rectangle26_6 {
	return fixed.Rectangle26_6{}
}

func (sb *LineBreakBox) Whitespace() bool {
	return true
}

func (sb *LineBreakBox) Image() image.Image {
	return image.NewRGBA(sb.ImageRect())
}

func (sb *LineBreakBox) ImageRect() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{},
	}
}
