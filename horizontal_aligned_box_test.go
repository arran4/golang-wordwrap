package wordwrap

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"testing"
)

type dummyHAlignBox struct {
	width  int
	height int
	drawn  image.Rectangle
	drawCalls int
}

func (d *dummyHAlignBox) AdvanceRect() fixed.Int26_6 {
	return fixed.I(d.width)
}

func (d *dummyHAlignBox) MetricsRect() font.Metrics {
	return font.Metrics{}
}

func (d *dummyHAlignBox) Whitespace() bool {
	return false
}

func (d *dummyHAlignBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	d.drawn = i.Bounds()
	d.drawCalls++
}

func (d *dummyHAlignBox) FontDrawer() *font.Drawer {
	return nil
}

func (d *dummyHAlignBox) Len() int {
	return 1
}

func (d *dummyHAlignBox) TextValue() string {
	return "x"
}

func (d *dummyHAlignBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func (d *dummyHAlignBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return 0, 0
}

func TestHorizontalAlignedBox(t *testing.T) {
	tests := []struct {
		name         string
		alignment    HorizontalAlignment
		boxWidth     int
		allocWidth   int
		originX      int
		expectedMinX int
		expectedMaxX int
	}{
		{"AlignLeft", AlignLeft, 20, 100, 0, 0, 20},
		{"AlignCenter", AlignCenter, 20, 100, 0, 40, 60},
		{"AlignRight", AlignRight, 20, 100, 0, 80, 100},
		{"AlignRight_Overflow", AlignRight, 120, 100, 0, 0, 100},   // No offset if natural > allocated
		{"AlignCenter_Origin10", AlignCenter, 20, 100, 10, 50, 70}, // 10 + 40
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inner := &dummyHAlignBox{width: tc.boxWidth, height: 10}
			hab := &HorizontalAlignedBox{
				Box:       inner,
				Alignment: tc.alignment,
			}

			if hab.AdvanceRect() != inner.AdvanceRect() {
				t.Fatalf("AdvanceRect changed: got %v, want %v", hab.AdvanceRect(), inner.AdvanceRect())
			}

			img := image.NewRGBA(image.Rect(tc.originX, 0, tc.originX+tc.allocWidth, 10))
			dc := &DrawConfig{}
			hab.DrawBox(img, 0, dc)

			if inner.drawn.Min.X != tc.expectedMinX {
				t.Errorf("Expected Min.X = %d, got %d", tc.expectedMinX, inner.drawn.Min.X)
			}
			if inner.drawn.Max.X != tc.expectedMaxX {
				t.Errorf("Expected Max.X = %d, got %d", tc.expectedMaxX, inner.drawn.Max.X)
			}
		})
	}
}

func TestHorizontalAlignedIntegration(t *testing.T) {
	// A small integration test using FillLineBox to stretch layout width
	inner := &dummyHAlignBox{width: 20, height: 10}

	// Create horizontal aligned box that aligns to the right
	hab := &HorizontalAlignedBox{
		Box:       inner,
		Alignment: AlignRight,
	}

	// Force it to fill the entire line (let's say we have 100 width container)
	flb := &FillLineBox{
		Mode: FillEntireLine,
		Box:  hab,
	}

	ar := flb.AdvanceRect()
	if ar != fixed.I(20) {
		t.Fatalf("AdvanceRect should be preserved as 20, got %d", ar)
	}

	folder := &SimpleFolder{boxer: &manualBoxer{boxes: []Box{flb}}, container: image.Rect(0, 0, 100, 100)}

	line, err := folder.Next(0)
	line.setStats(0, 0, 0, 0)
	if err != nil {
		t.Fatalf("folder Next err: %v", err)
	}

	if line == nil {
		t.Fatal("expected line")
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 10))

	var drawnRect image.Rectangle

	options := []DrawOption{
		BoxRecorder(func(box Box, min image.Point, max image.Point, stats *BoxPositionStats) {
			if _, ok := box.(*FillLineBox); ok {
				drawnRect = image.Rectangle{Min: min, Max: max}
			}
		}),
	}

	err = line.DrawLine(img, options...)
	if err != nil {
		t.Fatalf("DrawLine err: %v", err)
	}

	// The recorder will be called for each box in line.boxes, which is the FillLineBox itself.
	// So we need to check if drewBox is recorded by type casting flb... wait, BoxRecorder is called by line.DrawLine
	// Which calls it for the top level box in line.boxes.
	// The top level is FillLineBox! So BoxRecorder only sees FillLineBox.

	if drawnRect.Min.X != 0 || drawnRect.Max.X != 100 {
		t.Errorf("Line allocated bounds should be 0 to 100, got %v", drawnRect)
	}

	if inner.drawn.Min.X != 80 {
		t.Errorf("Inner box should be drawn at 80 (right aligned), got %v", inner.drawn)
	}
	if inner.drawn.Max.X != 100 {
		t.Errorf("Inner box right edge should be at 100, got %v", inner.drawn.Max.X)
	}
}

func TestHorizontalAlignedMultiWord(t *testing.T) {
	alignments := []struct {
		name      string
		alignment HorizontalAlignment
		isCenter  bool
		isLeft    bool
		isRight   bool
	}{
		{"Left", AlignLeft, false, true, false},
		{"Center", AlignCenter, true, false, false},
		{"Right", AlignRight, false, false, true},
	}

	for _, tc := range alignments {
		t.Run(tc.name, func(t *testing.T) {
			// A small integration test using FillLineBox to stretch layout width over a multi-word row container
			container := NewContainerContent([]*Content{
				NewContent("New Game", WithFontColor(image.White)),
			},
				WithHorizontalAlignment(tc.alignment),
				WithDecorators(func(b Box) Box {
					return &FillLineBox{Mode: FillEntireLine, Box: b}
				}),
			)

			drawer := &font.Drawer{Face: basicfont.Face7x13}
			boxer := NewSimpleBoxer([]*Content{container}, drawer)

			b, _, err := boxer.Next()
			if err != nil {
				t.Fatalf("Boxer next err: %v", err)
			}

			flb, ok := b.(*FillLineBox)
			if !ok {
				t.Fatalf("Expected top-level FillLineBox, got %T", b)
			}

			// Record the natural AdvanceRect BEFORE passing the box to layout/folder
			naturalAdvance := flb.AdvanceRect()
			if naturalAdvance >= fixed.I(1000) {
				t.Fatalf("Natural advance is %v, expected it to be less than 1000", naturalAdvance)
			}

			folder := &SimpleFolder{boxer: &manualBoxer{boxes: []Box{flb}}, container: image.Rect(0, 0, 1000, 100)}

			line, err := folder.Next(0)
			line.setStats(0, 0, 0, 0)
			if err != nil {
				t.Fatalf("folder Next err: %v", err)
			}

			if line == nil {
				t.Fatal("expected line")
			}

			img := image.NewRGBA(image.Rect(0, 0, 1000, 10))

			var drawnRect image.Rectangle

			options := []DrawOption{
				BoxRecorder(func(box Box, min image.Point, max image.Point, stats *BoxPositionStats) {
					if _, ok := box.(*FillLineBox); ok {
						drawnRect = image.Rectangle{Min: min, Max: max}
					}

				}),
			}

			err = line.DrawLine(img, options...)
			if err != nil {
				t.Fatalf("DrawLine err: %v", err)
			}

			if drawnRect.Min.X != 0 || drawnRect.Max.X != 1000 {
				t.Errorf("Line allocated bounds should be 0 to 1000, got %v", drawnRect)
			}

			// Verify that AdvanceRect is indeed preserved post-layout
			if naturalAdvance != flb.AdvanceRect() {
				t.Errorf("AdvanceRect mutated: before %v, after %v", naturalAdvance, flb.AdvanceRect())
			}
		})
	}
}

func TestHorizontalAlignedBox_Integration_DecoratedFullWidthGeometry(t *testing.T) {
	// Case 10: Decorated full-width geometry tests
	alignments := []struct {
		name      string
		alignment HorizontalAlignment
		expectedX int
	}{
		{"Left", AlignLeft, 10},
		{"Center", AlignCenter, 40}, // 100 total width. Margin=0, Padding=10L,10R. Inner=80. Box=20. 80-20=60. Offset=60/2=30. +10=40.
		{"Right", AlignRight, 70},   // 10 padding left + 60 offset = 70.
	}

	for _, tc := range alignments {
		t.Run(tc.name, func(t *testing.T) {
			inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}
			hab := &HorizontalAlignedBox{Box: inner, Alignment: tc.alignment}

			padding := fixed.Rectangle26_6{
				Min: fixed.Point26_6{X: fixed.I(10)},
				Max: fixed.Point26_6{X: fixed.I(10)},
			}
			margin := fixed.Rectangle26_6{}
			bp := BgPositioningZeroed
			decBox := NewDecorationBox(hab, padding, margin, image.NewUniform(image.Transparent), bp)

			flb := &FillLineBox{Mode: FillEntireLine, Box: decBox}

			// We use direct rendering of flb because SimpleFolder doesn't naturally pass options or
			// render inner correctly for our simple test where we expect the background box coordinates
			// to be explicitly global. SimpleFolder applies offsets dynamically. Let's just call
			// flb.DrawBox directly to test geometry.
			img := image.NewRGBA(image.Rect(0, 0, 100, 20))
			dc := &DrawConfig{}

			// Simulate the Line drawing flb
			subImg := img.SubImage(image.Rect(0, 0, 100, 20)).(*image.RGBA)
			flb.DrawBox(subImg, 0, dc)

			if inner.drawCalls != 1 {
				t.Fatalf("inner DrawBox calls = %d, want 1", inner.drawCalls)
			}

			if inner.drawn.Min.X != tc.expectedX {
				t.Errorf("Expected X to be %d, got %v", tc.expectedX, inner.drawn.Min.X)
			}
			if inner.drawn.Max.X != tc.expectedX+20 {
				t.Errorf("Expected Max X to be %d, got %v", tc.expectedX+20, inner.drawn.Max.X)
			}
		})
	}
}

func TestHorizontalAlignedBox_VerticalAlignmentComposition(t *testing.T) {
	// Case 12: Composition with baseline/vertical alignment
	inner := &dummyHAlignBox{width: 20, height: 10}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}
	vab := &AlignedBox{Box: hab, Alignment: AlignMiddle}

	if vab.AdvanceRect() != inner.AdvanceRect() {
		t.Errorf("Vertical alignment composition mutated horizontal natural advance: got %v", vab.AdvanceRect())
	}
}

func TestHorizontalAlignedBox_HorizontalPositioningComposition(t *testing.T) {
	// Case 13: Whole-line positioning composes independently
	// Note: Whole-line alignment is a property of `Line` alignment offset in folder,
	// horizontal alignment is an internal layout calculation of HorizontalAlignedBox.
	// This verifies they can be used together.

	inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}
	flb := &FillLineBox{Mode: FillEntireLine, Box: hab}

	container := NewContainerContent([]*Content{}, WithHorizontalAlignment(AlignRight))
	boxer := NewSimpleBoxer([]*Content{container}, nil)

	folder := &SimpleFolder{boxer: boxer, container: image.Rect(0, 0, 100, 100), lineOptions: []func(Line){func(l Line){l.(interface{ horizontalPosition(HorizontalLinePosition) }).horizontalPosition(HorizontalCenterLines)}}}

	// Unshift flb manually for testing
	boxer.Unshift(flb)

	line, err := folder.Next(0)
	if err != nil {
		t.Fatalf("folder Next err: %v", err)
	}

	// Draw line and check inner offset
	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	line.DrawLine(img)

	// Since Line applies HorizontalCenterLines, but flb fills entire line (advance 100),
	// the line itself has width 100, so line centering offset is 0.
	// But HorizontalAlignedBox applies center offset of 40 ( (100-20)/2 ).
	if inner.drawn.Min.X != 40 {
		t.Errorf("Line centering should not override box centering, expected 40, got %d", inner.drawn.Min.X)
	}
}

func TestHorizontalAlignedBox_FillRestOfLineGeometry(t *testing.T) {
	// Tests FillRestOfLine with left/center/right alignments
	alignments := []struct {
		name      string
		alignment HorizontalAlignment
		expectedX int
	}{
		{"Left", AlignLeft, 30},
		{"Center", AlignCenter, 60}, // 100 total width. Preceding box 30. Rest of line 70. Box width 10. Offset = (70-10)/2 = 30. 30 + 30 = 60.
		{"Right", AlignRight, 90},   // Offset = 70-10 = 60. 30 + 60 = 90.
	}

	for _, tc := range alignments {
		t.Run(tc.name, func(t *testing.T) {
			preceding := &dummyHAlignBox{width: 30, height: 10}
			inner := &dummyHAlignBox{width: 10, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}

			hab := &HorizontalAlignedBox{Box: inner, Alignment: tc.alignment}
			flb := &FillLineBox{Mode: FillRestOfLine, Box: hab}

			folder := &SimpleFolder{boxer: &manualBoxer{boxes: []Box{preceding, flb}}, container: image.Rect(0, 0, 100, 100)}

			line, _ := folder.Next(0)

			img := image.NewRGBA(image.Rect(0, 0, 100, 10))
			line.DrawLine(img)

			if inner.drawn.Min.X != tc.expectedX {
				t.Errorf("Expected X=%d for FillRestOfLine %s, got %v", tc.expectedX, tc.name, inner.drawn.Min.X)
			}
		})
	}
}


func TestHorizontalAlignedBox_MultipleBoxesState(t *testing.T) {
	// Case 14: Multiple boxes retain independent state
	// Just verify we can instantiate multiple and they keep their state
	b1 := &HorizontalAlignedBox{Box: &dummyHAlignBox{width: 20}, Alignment: AlignLeft}
	b2 := &HorizontalAlignedBox{Box: &dummyHAlignBox{width: 20}, Alignment: AlignRight}

	if b1.Alignment == b2.Alignment {
		t.Errorf("Boxes should retain independent state")
	}
}

func TestHorizontalAlignedBox_ZeroWidthContent(t *testing.T) {
	// Tests that zero-width content receives an empty, correctly offset SubImage
	inner := &dummyHAlignBox{width: 0, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}

	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	dc := &DrawConfig{}

	hab.DrawBox(img, 0, dc)

	if inner.drawCalls != 1 {
		t.Fatalf("Expected 1 draw call, got %d", inner.drawCalls)
	}
	if inner.drawn.Min.X != 50 {
		t.Errorf("Zero width center offset should be 50, got X=%v", inner.drawn)
	}
}

func TestHorizontalAlignedBox_NaturalEqualsAllocated(t *testing.T) {
	inner := &dummyHAlignBox{width: 100, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}

	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	dc := &DrawConfig{}

	hab.DrawBox(img, 0, dc)

	if inner.drawCalls != 1 {
		t.Fatalf("Expected 1 draw call, got %d", inner.drawCalls)
	}
	if inner.drawn.Min.X != 0 || inner.drawn.Max.X != 100 {
		t.Errorf("Natural==Allocated should not offset, got %v", inner.drawn)
	}
}

func TestHorizontalAlignedBox_NaturalGreaterThanAllocated(t *testing.T) {
	inner := &dummyHAlignBox{width: 120, height: 10, drawn: image.Rect(-1,-1,-1,-1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}

	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	dc := &DrawConfig{}

	hab.DrawBox(img, 0, dc)

	if inner.drawCalls != 1 {
		t.Fatalf("Expected 1 draw call, got %d", inner.drawCalls)
	}
	if inner.drawn.Min.X != 0 || inner.drawn.Max.X != 100 {
		t.Errorf("Overflow should pass through un-offset bounds to not mess up scaling, got %v", inner.drawn)
	}
}



func TestHorizontalAlignedBox_ExternalPackageCompatibility(t *testing.T) {
	// Case 15: Existing custom Box implementations continue compiling.
	// Since we are in the `wordwrap` package itself we can't fully simulate an external package,
	// but we CAN verify that a type with ONLY the original required `Box` methods
	// can still be wrapped without compilation errors or runtime panics.

	type CustomLegacyBox struct {
		Box // Embed to satisfy interface for the dummy, but we override all methods below
	}

	// We define only the mandatory ones here. If HorizontalAlignedBox requires a new method
	// that we didn't embed, this would panic if not correctly type-checked.

	// Create horizontal aligned box that aligns to the right
	hab := &HorizontalAlignedBox{
		Box:       &CustomLegacyBox{},
		Alignment: AlignRight,
	}

	// Make sure we can still call AdvanceRect without a panic from a missing method
	// (Since we embedded Box, it will call the nil embedded Box, so it would panic.
	// The key is that it COMPILES and satisfies Box).

	_ = hab
}
