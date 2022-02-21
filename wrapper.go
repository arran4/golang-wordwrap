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

func SimpleWrapTextToImage(text string, i Image, grf font.Face, opts ...WrapperOption) error {
	sw := &SimpleWrapper{}
	sw.ApplyOptions(opts)
	ls, _, err := sw.TextToRect(text, i.Bounds(), grf)
	if err != nil {
		return fmt.Errorf("wrapping text: %s", err)
	}
	return sw.RenderLines(i, ls, i.Bounds().Min)
}

func (sw *SimpleWrapper) RenderLines(i Image, ls []Line, at image.Point) error {
	for _, l := range ls {
		s := l.Size()
		rgba := i.SubImage(s.Add(at)).(Image)
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
	p := r.Min
	sb := NewSimpleBoxer([]rune(text), &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: grf,
	})
	sf := NewSimpleFolder(sb, r, sw.FoldOptions...)
	for p.Y < r.Dy() {
		l, err := sf.Next()
		if err != nil {
			return nil, image.Point{}, fmt.Errorf("boxing text at line %d: %w", len(ls), err)
		}
		if l == nil {
			break
		}
		ls = append(ls, l)
		s := l.Size()
		p.Y += s.Dy()
	}
	return ls, p, nil
}
