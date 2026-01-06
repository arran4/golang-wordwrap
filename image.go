// Copyright 2024 arran4
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	for x := s.Min.X; x < s.Max.X; x++ {
		i.Set(x, s.Min.Y, srci.At(x, s.Min.Y))
		i.Set(x, s.Max.Y-1, srci.At(x, s.Max.Y-1))
	}
	for y := s.Min.Y; y < s.Max.Y; y++ {
		i.Set(s.Min.X, y, srci.At(s.Min.X, y))
		i.Set(s.Max.X-1, y, srci.At(s.Max.X-1, y))
	}
}
