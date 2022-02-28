package wordwrap

import (
	"fmt"
	"golang.org/x/image/font"
	"image"
)

// SimpleWrapper quick and dirty wrapper.
type SimpleWrapper struct {
	folderOptions []FolderOption
	boxerOptions  []BoxerOption
	boxer         Boxer
	fontDrawer    *font.Drawer
}

// addFoldConfig allows passing down of FolderOption
func (sw *SimpleWrapper) addFoldConfig(option FolderOption) {
	sw.folderOptions = append(sw.folderOptions, option)
}

// SimpleWrapTextToImage all in one helper function to wrap text onto an image. Use image.Image's SubImage() to specify
// the exact location to render:
// 		SimpleWrapTextToImage("text", i.SubImage(image.Rect(30,30,400,400)), font)
func SimpleWrapTextToImage(text string, i Image, grf font.Face, opts ...WrapperOption) error {
	sw := NewSimpleWrapper(text, grf, opts...)
	ls, _, err := sw.TextToRect(i.Bounds())
	if err != nil {
		return fmt.Errorf("wrapping text: %s", err)
	}
	return sw.RenderLines(i, ls, i.Bounds().Min)
}

// NewSimpleWrapper creates a new wrapper. This function retains previous text position, useful for creating "pages."
func NewSimpleWrapper(text string, grf font.Face, opts ...WrapperOption) *SimpleWrapper {
	fontDrawer := &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: grf,
	}
	sw := &SimpleWrapper{
		fontDrawer: fontDrawer,
	}
	sw.ApplyOptions(opts...)
	sw.boxer = NewSimpleBoxer([]rune(text), fontDrawer, sw.boxerOptions...)
	return sw
}

// RenderLines draws the boxes for the given lines. on the image, starting at the specified point ignoring the original
// boundaries but maintaining the wrapping
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

// SimpleWrapTextToRect calculates and returns the position of each box and the image.Point it would end.
func SimpleWrapTextToRect(text string, r image.Rectangle, grf font.Face, opts ...WrapperOption) (*SimpleWrapper, []Line, image.Point, error) {
	sw := NewSimpleWrapper(text, grf, opts...)
	l, p, err := sw.TextToRect(r)
	return sw, l, p, err
}

// ApplyOptions allows the application of options to the SimpleWrapper (Such as new fonts, or turning on / off boxes.
func (sw *SimpleWrapper) ApplyOptions(opts ...WrapperOption) {
	for _, opt := range opts {
		opt.ApplyWrapperConfig(sw)
	}
}

// addBoxConfig Adds a constructor box config option to the boxer
func (sw *SimpleWrapper) addBoxConfig(bo BoxerOption) {
	sw.boxerOptions = append(sw.boxerOptions, bo)
}

// TextToRect calculates and returns the position of each box and the image.Point it would end.
func (sw *SimpleWrapper) TextToRect(r image.Rectangle) ([]Line, image.Point, error) {
	ls := make([]Line, 0)
	p := r.Min
	sf := NewSimpleFolder(sw.boxer, r, sw.fontDrawer, sw.folderOptions...)
	for (p.Y - r.Min.Y) <= r.Dy() {
		l, err := sf.Next(r.Dy() - (p.Y - r.Min.Y))
		if err != nil {
			return nil, image.Point{}, fmt.Errorf("boxing text at line %d: %w", len(ls), err)
		}
		if l == nil {
			break
		}
		s := l.Size()
		stop := false
		switch sf.yOverflow {
		case StrictBorders:
			// Handled elsewhere
		case DescentOverflow:
			if (p.Y - r.Min.Y + l.YValue()) > r.Dy() {
				sf.boxer.Push(l.Boxes()...)
				stop = true
			}
		case FullOverflowDuplicate:
			if (p.Y - r.Min.Y + s.Dy()) > r.Dy() {
				sf.boxer.Push(l.Boxes()...)
			}
		}
		if stop {
			break
		}
		ls = append(ls, l)
		p.Y += s.Dy()
	}
	if sf.pageBreakBox != nil && len(ls) > 0 && sf.boxer.HasNext() {
		if err := ls[len(ls)-1].PopSpaceFor(sf, r, &PageBreakBox{Box: sf.pageBreakBox}); err != nil {
			return nil, image.Point{}, err
		}
	}
	sw.fontDrawer = sf.lastFontDrawer
	return ls, p, nil
}

// HasNext are there any unprocessed bytes in the boxer
func (sw *SimpleWrapper) HasNext() bool {
	return sw.boxer.HasNext()
}
