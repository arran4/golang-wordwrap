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
	return sw.RenderLines(i, ls, i.Rect.Min)
}

func (sw *SimpleWrapper) RenderLines(i *image.RGBA, ls []Line, at image.Point) error {
	for _, l := range ls {
		s := l.Size()
		rgba := i.SubImage(s.Add(at)).(*image.RGBA)
		if err := l.DrawLine(rgba); err != nil {
			return fmt.Errorf("drawing text: %s", err)
		}
		at.Y += s.Dy()
	}
	return nil
}

func SimpleWrapTextToRect(text string, r image.Rectangle, grf font.Face, opts ...WrapperOption) (*SimpleWrapper, []Line, image.Point, error) {
	sw := &SimpleWrapper{}
	sw.ApplyOptions(opts)
	l, p, err := sw.TextToRect(text, r, grf)
	return sw, l, p, err
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
