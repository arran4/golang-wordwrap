package wordwrap

import (
	"image"
	"image/draw"
)

// Image because image.Image / draw.Image should really have SubImage as part of it.
type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}

// SourceImageMapper allows passing in of an option that will map the original input in some way
type SourceImageMapper func(image.Image) image.Image

// Apply installs the image source mapper
func (s SourceImageMapper) Apply(config *DrawConfig) {
	if config.SourceImageMapper != nil {
		orig := config.SourceImageMapper
		config.SourceImageMapper = func(i image.Image) image.Image {
			return s(orig(i))
		}
	}
	config.SourceImageMapper = s
}

// Interface enforcement
var _ DrawOption = (*SourceImageMapper)(nil)

// DrawConfig options for the drawer
type DrawConfig struct {
	SourceImageMapper SourceImageMapper
}

// DrawOption options applied and passed down the drawing functions
type DrawOption interface {
	Apply(*DrawConfig)
}

// DrawBox literally draws a simple box
func DrawBox(i draw.Image, s image.Rectangle, options ...DrawOption) {
	var srci image.Image = image.Black
	dc := &DrawConfig{}
	for _, option := range options {
		option.Apply(dc)
	}
	if dc.SourceImageMapper != nil {
		originalSrc := srci
		srci = dc.SourceImageMapper(originalSrc)
		defer func() {
			srci = originalSrc
		}()
	}
	for x := s.Min.X; x < s.Max.X; x++ {
		i.Set(x, s.Min.Y, srci.At(x, s.Min.Y))
		i.Set(x, s.Max.Y-1, srci.At(x, s.Max.Y-1))
	}
	for y := s.Min.Y; y < s.Max.Y; y++ {
		i.Set(s.Min.X, y, srci.At(s.Min.X, y))
		i.Set(s.Max.X-1, y, srci.At(s.Max.X-1, y))
	}
}
