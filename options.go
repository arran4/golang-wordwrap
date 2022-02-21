package wordwrap

import "log"

type FolderOption interface {
	WrapperOption
	ApplyFoldConfig(interface{})
}

type BoxerOption interface {
	WrapperOption
	FolderOption
	ApplyBoxConfig(interface{})
}

type WrapperOption interface {
	ApplyWrapperConfig(interface{})
}

type boxerOptionFunc func(interface{})

var _ BoxerOption = boxerOptionFunc(nil)
var _ FolderOption = boxerOptionFunc(nil)
var _ WrapperOption = boxerOptionFunc(nil)

type addBoxConfig interface {
	addBoxConfig(BoxerOption)
}

var _ addBoxConfig = (*SimpleLine)(nil)

func (b boxerOptionFunc) ApplyWrapperConfig(wr interface{}) {
	if fr, ok := wr.(addFoldConfig); ok {
		fr.addFoldConfig(b)
	} else {
		log.Printf("can't apply")
	}
}

func (b boxerOptionFunc) ApplyBoxConfig(br interface{}) {
	b(br)
}
func (b boxerOptionFunc) ApplyFoldConfig(fr interface{}) {
	if fr, ok := fr.(interface{ addBoxConfig(BoxerOption) }); ok {
		fr.addBoxConfig(b)
	} else {
		log.Printf("can't apply")
	}
}

type folderOptionFunc func(interface{})

var _ FolderOption = folderOptionFunc(nil)
var _ WrapperOption = folderOptionFunc(nil)

type addFoldConfig interface {
	addFoldConfig(FolderOption)
}

var _ addFoldConfig = (*SimpleWrapper)(nil)

func (f folderOptionFunc) ApplyWrapperConfig(wr interface{}) {
	if wr, ok := wr.(addFoldConfig); ok {
		wr.addFoldConfig(f)
	} else {
		log.Printf("can't apply")
	}
}

func (f folderOptionFunc) ApplyFoldConfig(fr interface{}) {
	f(fr)
}

type wrapperOptionFunc func(interface{})

var _ WrapperOption = wrapperOptionFunc(nil)

type addWrapperConfig interface {
	addWrapperConfig(WrapperOption)
}

func (f wrapperOptionFunc) ApplyWrapperConfig(fr interface{}) {
	f(fr)
}

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
