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

// LinePositionStats numbers to use for pin pointing location
type LinePositionStats struct {
	LineNumber    int
	PageBoxOffset int
	WordOffset    int
	PageNumber    int
}

// BoxPositionStats generates object of same name
func (lps *LinePositionStats) BoxPositionStats(numberInLine int) *BoxPositionStats {
	return &BoxPositionStats{
		LinePositionStats: lps,
		NumberInLine:      numberInLine,
		PageBoxOffset:     numberInLine + lps.PageBoxOffset,
		WordOffset:        numberInLine + lps.WordOffset,
	}
}

// BoxPositionStats Box position stats
type BoxPositionStats struct {
	*LinePositionStats
	NumberInLine  int
	PageBoxOffset int
	WordOffset    int
}

// BoxDrawMap allows the modification of boxes
type BoxDrawMap func(box Box, drawOps *DrawConfig, bps *BoxPositionStats) Box

// Apply installs the image source mapper
func (s BoxDrawMap) Apply(config *DrawConfig) {
	if config.BoxDrawMap != nil {
		orig := config.BoxDrawMap
		config.BoxDrawMap = func(box Box, drawOps *DrawConfig, bps *BoxPositionStats) Box {
			return s(orig(box, drawOps, bps), drawOps, bps)
		}
	}
	config.BoxDrawMap = s
}

// Interface enforcement
var _ DrawOption = (*BoxDrawMap)(nil)

// DrawConfig options for the drawer
type DrawConfig struct {
	SourceImageMapper SourceImageMapper
	BoxDrawMap        BoxDrawMap
}

// ApplyMap applies the box mapping function used for conditionally rendering or modifying the object being rendered
func (c *DrawConfig) ApplyMap(b Box, bps *BoxPositionStats) Box {
	if c.BoxDrawMap != nil {
		return c.BoxDrawMap(b, c, bps)
	}
	return b
}

// NewDrawConfig construct a draw config from DrawOptions
func NewDrawConfig(options ...DrawOption) *DrawConfig {
	dc := &DrawConfig{}
	for _, option := range options {
		option.Apply(dc)
	}
	return dc
}

// DrawOption options applied and passed down the drawing functions
type DrawOption interface {
	Apply(*DrawConfig)
}

// DrawBox literally draws a simple box
func DrawBox(i draw.Image, s image.Rectangle, dc *DrawConfig) {
	var srci image.Image = image.Black
	if dc.SourceImageMapper != nil {
		originalSrc := srci
		srci = dc.SourceImageMapper(originalSrc)
		defer func() {
			srci = originalSrc
		}()
	}
	// Top
	draw.Draw(i, image.Rectangle{Min: s.Min, Max: image.Point{X: s.Max.X, Y: s.Min.Y + 1}}, srci, s.Min, draw.Src)
	// Bottom
	draw.Draw(i, image.Rectangle{Min: image.Point{X: s.Min.X, Y: s.Max.Y - 1}, Max: s.Max}, srci, image.Point{X: s.Min.X, Y: s.Max.Y - 1}, draw.Src)
	// Left
	draw.Draw(i, image.Rectangle{Min: s.Min, Max: image.Point{X: s.Min.X + 1, Y: s.Max.Y}}, srci, s.Min, draw.Src)
	// Right
	draw.Draw(i, image.Rectangle{Min: image.Point{X: s.Max.X - 1, Y: s.Min.Y}, Max: s.Max}, srci, image.Point{X: s.Max.X - 1, Y: s.Min.Y}, draw.Src)
}
