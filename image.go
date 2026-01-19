package wordwrap

import (
	"image"
	"image/color"
	"image/draw"
)

// BoxRecorder allows recording of the box's position
type BoxRecorder func(box Box, min, max image.Point, bps *BoxPositionStats)

// Image because image.Image / draw.Image should really have SubImage as part of it.
type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}

// Tiled creates a TiledImage that tiles the source image
func Tiled(img image.Image) image.Image {
	return &TiledImage{img}
}

// TiledImage implements infinite tiling of a source image
type TiledImage struct {
	Src image.Image
}

func (t *TiledImage) ColorModel() color.Model { return t.Src.ColorModel() }
func (t *TiledImage) Bounds() image.Rectangle {
	// Infinite bounds conceptually, but we must return something.
	// We return a very large rectangle to simulate infinity for draw.Draw?
	// Or we return the Src bounds and expect the consumer to know?
	// Standard draw.Draw will clip to Bounds().
	// So we should return a very large rect.
	return image.Rect(-1e9, -1e9, 1e9, 1e9)
}
func (t *TiledImage) At(x, y int) color.Color {
	b := t.Src.Bounds()
	w, h := b.Dx(), b.Dy()
	// Euclidean modulo
	x = (x - b.Min.X) % w
	if x < 0 {
		x += w
	}
	y = (y - b.Min.Y) % h
	if y < 0 {
		y += h
	}
	return t.Src.At(b.Min.X+x, b.Min.Y+y)
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

// Apply installs the image source mapper
func (s BoxRecorder) Apply(config *DrawConfig) {
	if config.BoxRecorder != nil {
		orig := config.BoxRecorder
		config.BoxRecorder = func(box Box, min, max image.Point, bps *BoxPositionStats) {
			orig(box, min, max, bps)
			s(box, min, max, bps)
		}
	}
	config.BoxRecorder = s
}

// Interface enforcement
var _ DrawOption = (*SourceImageMapper)(nil)
var _ DrawOption = (*BoxRecorder)(nil)

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
	BoxRecorder       BoxRecorder
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
	for x := s.Min.X; x < s.Max.X; x++ {
		i.Set(x, s.Min.Y, srci.At(x, s.Min.Y))
		i.Set(x, s.Max.Y-1, srci.At(x, s.Max.Y-1))
	}
	for y := s.Min.Y; y < s.Max.Y; y++ {
		i.Set(s.Min.X, y, srci.At(s.Min.X, y))
		i.Set(s.Max.X-1, y, srci.At(s.Max.X-1, y))
	}
}
