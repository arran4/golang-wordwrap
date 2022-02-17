package wordwrap

import "log"

type FolderOption interface {
	ApplyFoldConfig(interface{})
}

type BoxerOption interface {
	FolderOption
	ApplyBoxConfig(interface{})
}

type boxerOptionFunc func(interface{})

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

func (f folderOptionFunc) ApplyFoldConfig(fr interface{}) {
	f(fr)
}

var BoxLine = folderOptionFunc(func(f interface{}) {
	switch f := f.(type) {
	case interface{ turnOnBox() }:
		f.turnOnBox()
	default:
		log.Printf("can't apply")
	}
})

var BoxBox = boxerOptionFunc(func(f interface{}) {
	switch f := f.(type) {
	case interface{ turnOnBox() }:
		f.turnOnBox()
	}
})
