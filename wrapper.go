package wordwrap

import (
	"fmt"
	"golang.org/x/image/font"
	"image"
)

type SimpleWrapper struct {
	FoldOptions []FolderOption
}

func (sw *SimpleWrapper) addFoldConfig(option FolderOption) {
	sw.FoldOptions = append(sw.FoldOptions, option)
}

func SimpleWrapTextToImage(text string, i *image.RGBA, grf font.Face, opts ...WrapperOption) error {
	sw := &SimpleWrapper{}
	sw.ApplyOptions(opts)
	ls, _, err := sw.TextToRect(text, i.Rect, grf)
	if err != nil {
		return fmt.Errorf("wrapping text: %s", err)
	}
	return sw.TextToImage(i, ls)
}

func (sw *SimpleWrapper) TextToImage(i *image.RGBA, ls []Line) error {
	p := i.Rect.Min
	for _, l := range ls {
		s := l.Size()
		rgba := i.SubImage(s.Add(p)).(*image.RGBA)
		if err := l.DrawLine(rgba); err != nil {
			return fmt.Errorf("drawing text: %s", err)
		}
		p.Y += s.Dy()
	}
	return nil
}

func SimpleWrapTextToRect(text string, r image.Rectangle, grf font.Face, opts ...WrapperOption) ([]Line, image.Point, error) {
	sw := &SimpleWrapper{}
	sw.ApplyOptions(opts)
	return sw.TextToRect(text, r, grf)
}

func (sw *SimpleWrapper) ApplyOptions(opts []WrapperOption) {
	for _, opt := range opts {
		opt.ApplyWrapperConfig(sw)
	}
}

func (sw *SimpleWrapper) TextToRect(text string, r image.Rectangle, grf font.Face) ([]Line, image.Point, error) {
	ls := make([]Line, 0)
	n := 0
	rt := []rune(text)
	p := r.Min
	for p.Y < r.Dy() {
		l, ni, err := SimpleFolder(SimpleBoxer, grf, rt[n:], r, sw.FoldOptions...)
		if err != nil {
			return nil, image.Point{}, fmt.Errorf("boxing text at pos %d: %w", n, err)
		}
		if l == nil {
			break
		}
		ls = append(ls, l)
		n += ni
		s := l.Size()
		p.Y += s.Dy()
	}
	return ls, p, nil
}
