package wordwrap

import (
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

/*

At a later stage I'm going to change how this works. Currently, it isn't great but open to suggestions.

*/

// FolderOption for folders
type FolderOption interface {
	// WrapperOption Allows you to pass the option to a Wrapper and assume it gets passed to the constructor of the
	// Folder
	WrapperOption
	// ApplyFoldConfig applies the config.
	ApplyFoldConfig(interface{})
}

// BoxerOption for folders
type BoxerOption interface {
	// WrapperOption Allows you to pass the option to a Wrapper and assume it gets passed to the constructor of the
	// Boxer
	WrapperOption
	// ApplyBoxConfig applies the config.
	ApplyBoxConfig(interface{})
}

// WrapperOption for folders
type WrapperOption interface {
	// ApplyWrapperConfig applies the config.
	ApplyWrapperConfig(interface{})
}

// boxerOptionFunc Wrapper function that automatically fills: WrapperOption requirement for the FolderOption interface
type boxerOptionFunc func(interface{})

// Reports interface adherence
var _ BoxerOption = boxerOptionFunc(nil)
var _ WrapperOption = boxerOptionFunc(nil)

// addBoxConfig interface that applies the config
type addBoxConfig interface {
	addBoxConfig(BoxerOption)
}

// Reports interface adherence
var _ addBoxConfig = (*SimpleWrapper)(nil)

// ApplyWrapperConfig function that fills the: WrapperOption requirement for the BoxerOption interface
func (b boxerOptionFunc) ApplyWrapperConfig(wr interface{}) {
	if wr, ok := wr.(interface{ addBoxConfig(BoxerOption) }); ok {
		wr.addBoxConfig(b)
	} else {
		log.Printf("can't apply")
	}
}

// ApplyBoxConfig Converts function into an object to match interface
func (b boxerOptionFunc) ApplyBoxConfig(br interface{}) {
	b(br)
}

// folderOptionFunc Wrapper function that automatically fills: WrapperOption requirement for the FolderOption interface
type folderOptionFunc func(interface{})

// Reports interface adherence
var _ FolderOption = folderOptionFunc(nil)
var _ WrapperOption = folderOptionFunc(nil)

// addFoldConfig interface that applies the config
type addFoldConfig interface {
	addFoldConfig(FolderOption)
}

// Reports interface adherence
var _ addFoldConfig = (*SimpleWrapper)(nil)

// ApplyWrapperConfig function that fills the: WrapperOption requirement for the FolderOption interface
func (f folderOptionFunc) ApplyWrapperConfig(wr interface{}) {
	if wr, ok := wr.(addFoldConfig); ok {
		wr.addFoldConfig(f)
	} else {
		log.Printf("can't apply")
	}
}

// ApplyFoldConfig Converts function into an object to match interface
func (f folderOptionFunc) ApplyFoldConfig(fr interface{}) {
	f(fr)
}

// wrapperOptionFunc that converts a function into a WrapperOption interface
type wrapperOptionFunc func(interface{})

// Reports interface adherence
var _ WrapperOption = wrapperOptionFunc(nil)

// Commented out until used because of... linter.
//type addWrapperConfig interface {
//	addWrapperConfig(WrapperOption)
//}

// ApplyWrapperConfig Converts function into an object to match interface
func (f wrapperOptionFunc) ApplyWrapperConfig(fr interface{}) {
	f(fr)
}

// BoxLine is a FolderOption that tells the Liner to draw a box around the line mostly for debugging purposes but will be
// the basis of how select and highlighting could work
var BoxLine = folderOptionFunc(func(f interface{}) {
	if f, ok := f.(*SimpleFolder); ok {
		f.lineOptions = append(f.lineOptions, func(line Line) {
			switch line := line.(type) {
			case interface{ turnOnBox() }:
				line.turnOnBox()
			default:
				log.Printf("can't apply")
			}
		})
	}
})

// NewPageBreakBox is a FolderOption that tells the Liner to add a chevron image to the end of every text block that continues
// past the given rect.
func NewPageBreakBox(b Box, opts ...BoxerOption) WrapperOption {
	return folderOptionFunc(func(f interface{}) {
		for _, o := range opts {
			o.ApplyBoxConfig(b)
		}
		switch f := f.(type) {
		case interface{ setPageBreakBox(b Box) }:
			f.setPageBreakBox(b)
		default:
			log.Printf("can't apply")
		}
	})
}

// YOverflow is a FolderOption that sets the type over overflow mode we will allow
func YOverflow(i OverflowMode) WrapperOption {
	return folderOptionFunc(func(f interface{}) {
		if f, ok := f.(*SimpleFolder); ok {
			f.yOverflow = i
		}
	})
}

// BoxBox is a BoxerOption that tells the Box to draw a box around itself mostly for debugging purposes but will be
// the basis of how select and highlighting could work, such as the cursor
var BoxBox = boxerOptionFunc(func(f interface{}) {
	bf := func(box Box) {
		switch box := box.(type) {
		case interface{ turnOnBox() }:
			box.turnOnBox()
		}
	}
	switch f := f.(type) {
	case *SimpleBoxer:
		f.postBoxOptions = append(f.postBoxOptions, bf)

	case Box:
		bf(f)
	}
})

// ImageBoxOption modifiers for the ImageBox
type ImageBoxOption interface {
	applyImageBoxOption(box *ImageBox)
}

type imageBoxOptionMetricCalcFunc func(ib2 *ImageBox) font.Metrics

func (i imageBoxOptionMetricCalcFunc) applyImageBoxOption(box *ImageBox) {
	box.metricCalc = i
}

var _ ImageBoxOption = (imageBoxOptionMetricCalcFunc)(nil)

// ImageBoxMetricAboveTheLine Puts the image above the baseline as you would expect if you were using a word processor
var ImageBoxMetricAboveTheLine imageBoxOptionMetricCalcFunc = func(ib *ImageBox) font.Metrics {
	return font.Metrics{
		Height: fixed.I(ib.I.Bounds().Dy()),
		Ascent: fixed.I(ib.I.Bounds().Dy()),
	}
}

// ImageBoxMetricBelowTheLine Puts the image above the baseline. Rarely done
var ImageBoxMetricBelowTheLine imageBoxOptionMetricCalcFunc = func(ib *ImageBox) font.Metrics {
	return font.Metrics{
		Height:  fixed.I(ib.I.Bounds().Dy()),
		Descent: fixed.I(ib.I.Bounds().Dy()),
	}
}

// ImageBoxMetricCenter Puts the image running from the top down
var ImageBoxMetricCenter = func(fd *font.Drawer) imageBoxOptionMetricCalcFunc {
	return func(ib *ImageBox) font.Metrics {
		if fd == nil {
			fd = ib.fontDrawer
		}
		if fd == nil {
			return ImageBoxMetricBelowTheLine(ib)
		}
		m := fd.Face.Metrics()
		return font.Metrics{
			Height:  fixed.I(ib.I.Bounds().Dy()),
			Descent: fixed.I(ib.I.Bounds().Dy())/2 - m.Descent/2,
			Ascent:  fixed.I(ib.I.Bounds().Dy())/2 + m.Descent/2,
		}
	}
}

// FontDrawer a wrapper around *font.Draw used to set the font
type FontDrawer struct {
	d *font.Drawer
}

// NewFontDrawer a wrapper around *font.Draw used to set the font mostly for image
func NewFontDrawer(d *font.Drawer) *FontDrawer {
	return &FontDrawer{
		d: d,
	}
}

// applyImageBoxOption
func (f *FontDrawer) applyImageBoxOption(box *ImageBox) {
	box.fontDrawer = f.d
}

var (
	// Enforce interface adherence
	_ ImageBoxOption = (*FontDrawer)(nil)
)

// HorizontalLinePosition is the type for per-line level alignment.
type HorizontalLinePosition int

const (
	// LeftLines default, produces lines that are individually left justified.
	LeftLines HorizontalLinePosition = iota
	// HorizontalCenterLines produces lines that are individually center justified.
	HorizontalCenterLines
	// RightLines produces lines that are individually right justified.
	RightLines
)

var (
	// Ensures interface compliance
	_ FolderOption = LeftLines
	// Ensures interface compliance
	_ WrapperOption = LeftLines
)

// ApplyWrapperConfig Is required to pass the configuration through to the appropriate level -- Hopefully will be
// refactored
func (hp HorizontalLinePosition) ApplyWrapperConfig(wr interface{}) {
	if wr, ok := wr.(addFoldConfig); ok {
		wr.addFoldConfig(hp)
	} else {
		log.Printf("can't apply")
	}
}

// ApplyFoldConfig applies the configuration to the wrapper where it will be stored in the line.
func (hp HorizontalLinePosition) ApplyFoldConfig(f interface{}) {
	if f, ok := f.(*SimpleFolder); ok {
		f.lineOptions = append(f.lineOptions, func(line Line) {
			switch line := line.(type) {
			case interface{ horizontalPosition(HorizontalLinePosition) }:
				line.horizontalPosition(hp)
			default:
				log.Printf("can't apply")
			}
		})
	}
}

// HorizontalBlockPosition information about how to position the entire block of text rather than just the line horizontally
type HorizontalBlockPosition int

const (
	// LeftBLock positions the entire block of text left rather than just the line horizontally (default)
	LeftBLock HorizontalBlockPosition = iota
	// HorizontalCenterBlock positions the entire block of text center rather than just the line horizontally
	HorizontalCenterBlock
	// RightBlock positions the entire block of text right rather than just the line horizontally
	RightBlock
)

// Interface enforcement
var _ WrapperOption = LeftBLock

// ApplyWrapperConfig Stores the position against the wrapper object
func (hp HorizontalBlockPosition) ApplyWrapperConfig(wr interface{}) {
	switch block := wr.(type) {
	case interface{ horizontalPosition(HorizontalBlockPosition) }:
		block.horizontalPosition(hp)
	default:
		log.Printf("can't apply")
	}
}

// VerticalBlockPosition information about how to position the entire block of text rather than just the line vertically
type VerticalBlockPosition int

const (
	// TopBLock positions the entire block of text top
	TopBLock VerticalBlockPosition = iota
	// VerticalCenterBlock positions the entire block of text center
	VerticalCenterBlock
	// BottomBlock positions the entire block of text bottom
	BottomBlock
)

// Interface enforcement
var _ WrapperOption = TopBLock

// ApplyWrapperConfig Stores the position against the wrapper object
func (hp VerticalBlockPosition) ApplyWrapperConfig(wr interface{}) {
	switch block := wr.(type) {
	case interface{ verticalPosition(VerticalBlockPosition) }:
		block.verticalPosition(hp)
	default:
		log.Printf("can't apply")
	}
}
