package wordwrap

import (
	"log"
)

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
func NewPageBreakBox(b Box) WrapperOption {
	return folderOptionFunc(func(f interface{}) {
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
	if f, ok := f.(*SimpleBoxer); ok {
		f.postBoxOptions = append(f.postBoxOptions, func(box Box) {
			switch box := box.(type) {
			case interface{ turnOnBox() }:
				box.turnOnBox()
			}
		})
	}
})
