package wordwrap

import (
	"fmt"
	"golang.org/x/image/font"
	"image"
)

type SimpleWrapper struct {
	folderOptions []FolderOption
	boxerOptions  []BoxerOption
	sb            Boxer
}

func (sw *SimpleWrapper) addFoldConfig(option FolderOption) {
	sw.folderOptions = append(sw.folderOptions, option)
}

func SimpleWrapTextToImage(text string, i Image, grf font.Face, opts ...WrapperOption) error {
	sw := NewSimpleWrapper(text, grf, opts...)
	ls, _, err := sw.TextToRect(i.Bounds())
	if err != nil {
		return fmt.Errorf("wrapping text: %s", err)
	}
	return sw.RenderLines(i, ls, i.Bounds().Min)
}

func NewSimpleWrapper(text string, grf font.Face, opts ...WrapperOption) *SimpleWrapper {
	sw := &SimpleWrapper{}
	sw.sb = NewSimpleBoxer([]rune(text), &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: grf,
	}, sw.boxerOptions...)
	sw.ApplyOptions(opts)
	return sw
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
	sw := NewSimpleWrapper(text, grf, opts...)
	l, p, err := sw.TextToRect(r)
	return sw, l, p, err
}

func (sw *SimpleWrapper) ApplyOptions(opts []WrapperOption) {
	for _, opt := range opts {
		opt.ApplyWrapperConfig(sw)
	}
}

func (sw *SimpleWrapper) addBoxConfig(bo BoxerOption) {
	sw.boxerOptions = append(sw.boxerOptions, bo)
}

func (sw *SimpleWrapper) TextToRect(r image.Rectangle) ([]Line, image.Point, error) {
	ls := make([]Line, 0)
	p := r.Min
	sf := NewSimpleFolder(sw.sb, r, sw.folderOptions...)
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
