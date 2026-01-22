package wordwrap

import (
	"fmt"
	"image"

	"golang.org/x/image/font"
)

// Deprecated: Moved to github.com/arran4/golang-rich-text/richtext
// SimpleWrapper provides basic text wrapping functionality.
type SimpleWrapper struct {
	folderOptions           []FolderOption
	boxerOptions            []BoxerOption
	boxer                   Boxer
	fontDrawer              *font.Drawer
	currentPage             int
	boxCount                int
	horizontalBlockPosition HorizontalBlockPosition
	verticalBlockPosition   VerticalBlockPosition
}

// horizontalPosition sets the horizontalBlockPosition
func (sw *SimpleWrapper) horizontalPosition(hp HorizontalBlockPosition) {
	sw.horizontalBlockPosition = hp
}

// verticalPosition sets the verticalBlockPosition
func (sw *SimpleWrapper) verticalPosition(hp VerticalBlockPosition) {
	sw.verticalBlockPosition = hp
}

// addFoldConfig allows passing down of FolderOption
func (sw *SimpleWrapper) addFoldConfig(option FolderOption) {
	sw.folderOptions = append(sw.folderOptions, option)
}

// SimpleWrapTextToImage all in one helper function to wrap text onto an image. Use image.Image's SubImage() to specify
// the exact location to render:
//
//	SimpleWrapTextToImage("text", i.SubImage(image.Rect(30,30,400,400)), font)
func SimpleWrapTextToImage(text string, i Image, grf font.Face, opts ...WrapperOption) error {
	sw := NewSimpleWrapper([]*Content{{text: text}}, grf, opts...)
	ls, _, err := sw.TextToRect(i.Bounds())
	if err != nil {
		return fmt.Errorf("wrapping text: %s", err)
	}
	return sw.RenderLines(i, ls, i.Bounds().Min)
}

// NewSimpleWrapper creates a new wrapper. This function retains previous text position, useful for creating "pages."
// assumes black text
//
// Deprecated: Moved to github.com/arran4/golang-rich-text/richtext
func NewSimpleWrapper(contents []*Content, grf font.Face, opts ...WrapperOption) *SimpleWrapper {
	args := []interface{}{contents, grf}
	for _, opt := range opts {
		args = append(args, opt)
	}
	return NewRichWrapper(args...)
}

// NewRichWrapper creates a new wrapper. valid args are font.Face, string, and WrapperOption
func NewRichWrapper(args ...interface{}) *SimpleWrapper {
	contents, fontDrawer, wrapperOptions, boxerOptions, boxer, tokenizer := ProcessRichArgs(args...)

	sw := &SimpleWrapper{
		fontDrawer:   fontDrawer,
		boxerOptions: boxerOptions,
	}
	sw.ApplyOptions(wrapperOptions...)
	if boxer == nil {
		sb := NewSimpleBoxer(contents, fontDrawer, sw.boxerOptions...)
		if tokenizer != nil {
			sb.Tokenizer = tokenizer
		}
		sw.boxer = sb
	} else {
		sw.boxer = boxer
	}
	return sw
}

// HorizontalLinePositioner is a simple interface denoting a getter
type HorizontalLinePositioner interface {
	GetHorizontalLinePosition() HorizontalLinePosition
}

// RenderLines draws the boxes for the given lines. on the image, starting at the specified point ignoring the original
// boundaries but maintaining the wrapping. Also applies alignment options.
func (sw *SimpleWrapper) RenderLines(i Image, ls []Line, at image.Point, options ...DrawOption) error {
	bounds := i.Bounds()
	offset := sw.calculateAlignmentOffset(ls, bounds)
	for _, l := range ls {
		s := l.Size()
		if l, ok := l.(HorizontalLinePositioner); ok {
			switch l.GetHorizontalLinePosition() {
			case HorizontalCenterLines:
				s = s.Add(image.Pt((bounds.Max.X-(s.Max.X-s.Min.X))/2, 0))
			case RightLines:
				s = s.Add(image.Pt(bounds.Max.X-(s.Max.X-s.Min.X), 0))
			}
		}
		rgba := i.SubImage(s.Add(offset).Add(at)).(Image)
		if err := l.DrawLine(rgba, options...); err != nil {
			return fmt.Errorf("drawing text: %s", err)
		}
		at.Y += s.Dy()
	}
	return nil
}

// calculateAlignmentOffset calculates the appropriate alignment offset for the block alignments VerticalBlockPosition
// and HorizontalBlockPosition
func (sw *SimpleWrapper) calculateAlignmentOffset(ls []Line, bounds image.Rectangle) (offset image.Point) {
	var actualSize *image.Point
	if sw.horizontalBlockPosition != LeftBLock || sw.verticalBlockPosition != TopBLock {
		actualSize = &image.Point{}
		for _, l := range ls {
			s := l.Size()
			actualSize.Y += s.Dy()
			if s.Dx() > actualSize.X {
				actualSize.X = s.Dx()
			}
		}
	}
	switch sw.horizontalBlockPosition {
	case HorizontalCenterBlock:
		offset.X = (bounds.Dx() - actualSize.X) / 2
	case RightBlock:
		offset.X = bounds.Dx() - actualSize.X
	}
	switch sw.verticalBlockPosition {
	case VerticalCenterBlock:
		offset.Y = (bounds.Dy() - actualSize.Y) / 2
	case BottomBlock:
		offset.Y = bounds.Dy() - actualSize.Y
	}
	return offset
}

// SimpleWrapTextToRect calculates and returns the position of each box and the image.Point it would end.
func SimpleWrapTextToRect(text string, r image.Rectangle, grf font.Face, opts ...WrapperOption) (*SimpleWrapper, []Line, image.Point, error) {
	sw := NewSimpleWrapper([]*Content{{text: text}}, grf, opts...)
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

type FitterConfig struct {
	IgnoreY bool
}

type FitterOption interface {
	Apply(*FitterConfig)
}

type FitterIgnoreY struct{}

func (fiy FitterIgnoreY) Apply(c *FitterConfig) {
	c.IgnoreY = true
}

var _ FitterOption = (*FitterIgnoreY)(nil)

// TextToRect calculates and returns the position of each box and the image.Point it would end.
func (sw *SimpleWrapper) TextToRect(r image.Rectangle, ops ...FitterOption) ([]Line, image.Point, error) {
	config := FitterConfig{}
	for _, op := range ops {
		op.Apply(&config)
	}
	ls := make([]Line, 0)
	p := r.Min
	sf := NewSimpleFolder(sw.boxer, r, sw.fontDrawer, sw.folderOptions...)
	pageBoxCount := 0
	for (p.Y-r.Min.Y) <= r.Dy() || config.IgnoreY {
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
		l.SetStats(len(ls), sw.currentPage, sw.boxCount, pageBoxCount)
		boxCount := len(l.Boxes())
		sw.boxCount += boxCount
		pageBoxCount += boxCount
		ls = append(ls, l)
		p.Y += s.Dy()
	}
	if sf.pageBreakBox != nil && sf.boxer.HasNext() {
		if len(ls) > 0 {
			line := ls[len(ls)-1]
			if n, err := line.PopSpaceFor(sf, r, NewPageBreak(sf.pageBreakBox)); err != nil {
				return nil, image.Point{}, err
			} else {
				sw.boxCount -= n
			}
			if len(line.Boxes()) == 1 {
				return nil, image.Point{}, fmt.Errorf("page break too long or rect too small")
			}
		} else if sf.pageBreakBox.MetricsRect().Height.Ceil() > r.Dy() {
			return nil, image.Point{}, fmt.Errorf("page break too tall or rect too small")
		}
	}
	sw.currentPage++
	sw.fontDrawer = sf.lastFontDrawer
	return ls, p, nil
}

// HasNext are there any unprocessed bytes in the boxer
func (sw *SimpleWrapper) HasNext() bool {
	return sw.boxer.HasNext()
}
