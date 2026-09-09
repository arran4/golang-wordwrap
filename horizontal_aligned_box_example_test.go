package wordwrap_test

import (
	"fmt"
	"image"

	"github.com/arran4/golang-wordwrap"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// simpleWidthBox is a minimal box for examples with a fixed 20px advance
type simpleWidthBox struct{}

func (s *simpleWidthBox) AdvanceRect() fixed.Int26_6              { return fixed.I(20) }
func (s *simpleWidthBox) MetricsRect() font.Metrics               { return font.Metrics{} }
func (s *simpleWidthBox) Whitespace() bool                        { return false }
func (s *simpleWidthBox) FontDrawer() *font.Drawer                { return nil }
func (s *simpleWidthBox) Len() int                                { return 1 }
func (s *simpleWidthBox) TextValue() string                       { return "X" }
func (s *simpleWidthBox) MinSize() (fixed.Int26_6, fixed.Int26_6) { return 0, 0 }
func (s *simpleWidthBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) { return 0, 0 }
func (s *simpleWidthBox) DrawBox(i wordwrap.Image, y fixed.Int26_6, dc *wordwrap.DrawConfig) {
	fmt.Printf("Drawn Box Bounds: %v\n", i.Bounds())
}

func ExampleHorizontalAlignedBox_center() {
	// A box with natural width of 20px
	inner := &simpleWidthBox{}

	// Tell it to center itself inside its parent layout allocation
	hab := &wordwrap.HorizontalAlignedBox{
		Box:       inner,
		Alignment: wordwrap.AlignCenter,
	}

	// Create a dummy image and draw config
	img := image.NewRGBA(image.Rect(0, 0, 100, 20))
	dc := &wordwrap.DrawConfig{}

	// Directly supply a 100px allocated SubImage to HorizontalAlignedBox.
	// (In real usage, this allocation is provided by folding components like FillLineBox)
	// Center alignment offset: (100 allocated - 20 natural) / 2 = 40.
	// So inner receives bounds offset by 40: (40, 0) - (60, 20).
	subImg := img.SubImage(image.Rect(0, 0, 100, 20)).(wordwrap.Image)
	hab.DrawBox(subImg, 0, dc)

	// Output:
	// Drawn Box Bounds: (40,0)-(60,20)
}

func ExampleHorizontalAlignedBox_right() {
	// A box with natural width of 20px
	inner := &simpleWidthBox{}

	// Tell it to right-align itself inside its parent layout allocation
	hab := &wordwrap.HorizontalAlignedBox{
		Box:       inner,
		Alignment: wordwrap.AlignRight,
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 20))
	dc := &wordwrap.DrawConfig{}

	// Directly supply a 100px allocated SubImage to HorizontalAlignedBox.
	// (In real usage, this allocation is provided by folding components like FillLineBox)
	// Right alignment offset: (100 allocated - 20 natural) = 80.
	// So inner receives bounds offset by 80: (80, 0) - (100, 20).
	subImg := img.SubImage(image.Rect(0, 0, 100, 20)).(wordwrap.Image)
	hab.DrawBox(subImg, 0, dc)

	// Output:
	// Drawn Box Bounds: (80,0)-(100,20)
}
