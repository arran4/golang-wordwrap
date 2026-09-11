package wordwrap

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"testing"
)

type dummyHAlignBox struct {
	width     int
	height    int
	drawn     image.Rectangle
	drawCalls int
	lastY     fixed.Int26_6
	fracWidth fixed.Int26_6
}

func (d *dummyHAlignBox) AdvanceRect() fixed.Int26_6 {
	if d.fracWidth != 0 {
		return d.fracWidth
	}
	return fixed.I(d.width)
}

func (d *dummyHAlignBox) MetricsRect() font.Metrics {
	if d.height > 0 {
		return font.Metrics{Ascent: fixed.I(d.height), Descent: 0}
	}
	return font.Metrics{}
}

func (d *dummyHAlignBox) Whitespace() bool {
	return false
}

func (d *dummyHAlignBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	d.drawn = i.Bounds()
	d.drawCalls++
	d.lastY = y
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

func TestHorizontalAlignedBox_HorizontalPositioningComposition(t *testing.T) {
	// Case 13: Whole-line positioning composes independently
	// We construct a line of 60px width inside a 100px render target.
	// We center the line itself (+20px offset to line).
	// Within the line, we have a box allocated 60px but naturally 20px, aligned right (+40px offset).
	// Final expected X = 20 + 40 = 60.

	inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignRight}

	// We can manually provide an allocated width wider than the natural width without filling the entire container.
	// To do so we just use an inline dummy Box with large AdvanceRect and wrap it, or just use a container size of 60
	// for the Folder layout, and then RenderLines into a 100px wide image!
	// Yes! Folder container = 60. So FillEntireLine will allocate exactly 60!

	flb := &FillLineBox{Mode: FillEntireLine, Box: hab}

	boxer := &manualBoxer{boxes: []Box{flb}}
	folder := &SimpleFolder{boxer: boxer, container: image.Rect(0, 0, 60, 100)} // 60 width container
	line, err := folder.Next(0)
	if err != nil {
		t.Fatalf("folder Next err: %v", err)
	}
	if line == nil {
		t.Fatalf("expected line")
	}

	sw := &SimpleWrapper{}
	img := image.NewRGBA(image.Rect(0, 0, 100, 10)) // 100 width render target

	line.(interface{ horizontalPosition(HorizontalLinePosition) }).horizontalPosition(HorizontalCenterLines)
	err = sw.RenderLines(img, []Line{line}, img.Bounds().Min)
	if err != nil {
		t.Fatalf("RenderLines err: %v", err)
	}

	// Line centering: (100 - 60) / 2 = 20
	// Box right align: 60 allocated - 20 natural = 40.
	// Total X = 20 + 40 = 60.
	if inner.drawn.Min.X != 60 {
		t.Errorf("Composition failed: expected total offset 60, got %d", inner.drawn.Min.X)
	}
}

func TestHorizontalAlignedBox_VerticalAlignmentComposition(t *testing.T) {
	// Case 12: Composition with baseline/vertical alignment
	inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
	hab := &HorizontalAlignedBox{Box: inner, Alignment: AlignCenter}
	vab := &AlignedBox{Box: hab, Alignment: AlignBottom}

	if vab.AdvanceRect() != inner.AdvanceRect() {
		t.Errorf("Vertical alignment composition mutated horizontal natural advance: got %v", vab.AdvanceRect())
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 20))
	dc := &DrawConfig{}

	vab.DrawBox(img, fixed.I(10), dc)

	if inner.drawn.Min.X != 40 {
		t.Errorf("Horizontal composition inside Vertical alignment failed, expected X=40, got %v", inner.drawn.Min.X)
	}

	// Our dummy has Ascent=10, Descent=0. Total height = 10.
	// `AlignBottom` changes Ascent=0, Descent=10.
	// innerY = 10 - 0 + 10 = 20.
	if inner.lastY != fixed.I(20) {
		t.Errorf("Vertical composition unexpectedly modified Y baseline: expected 20:00, got %v", inner.lastY)
	}
}

func TestHorizontalAlignedBox_FillEntireLineGeometry(t *testing.T) {
	// Tests FillEntireLine with left/center/right alignments folded dynamically
	alignments := []struct {
		name      string
		alignment HorizontalAlignment
		expectedX int
	}{
		{"Left", AlignLeft, 0},
		{"Center", AlignCenter, 40}, // 100 total width. Box width 20. Offset = (100-20)/2 = 40.
		{"Right", AlignRight, 80},   // Offset = 100-20 = 80.
	}

	for _, tc := range alignments {
		t.Run(tc.name, func(t *testing.T) {
			inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}

			hab := &HorizontalAlignedBox{Box: inner, Alignment: tc.alignment}
			flb := &FillLineBox{Mode: FillEntireLine, Box: hab}

			folder := &SimpleFolder{boxer: &manualBoxer{boxes: []Box{flb}}, container: image.Rect(0, 0, 100, 100)}

			line, err := folder.Next(0)
			if err != nil {
				t.Fatalf("folder Next err: %v", err)
			}
			if line == nil {
				t.Fatalf("expected line")
			}

			img := image.NewRGBA(image.Rect(0, 0, 100, 10))
			err = line.DrawLine(img)
			if err != nil {
				t.Fatalf("DrawLine err: %v", err)
			}

			if inner.drawn.Min.X != tc.expectedX {
				t.Errorf("Expected X=%d for FillEntireLine %s, got %v", tc.expectedX, tc.name, inner.drawn.Min.X)
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
			inner := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
			hab := &HorizontalAlignedBox{Box: inner, Alignment: tc.alignment}

			padding := fixed.Rectangle26_6{
				Min: fixed.Point26_6{X: fixed.I(10)},
				Max: fixed.Point26_6{X: fixed.I(10)},
			}
			margin := fixed.Rectangle26_6{}
			bp := BgPositioningZeroed

			// Use a non-transparent sentinel background to prove span.
			sentinelColor := image.NewUniform(image.Black)
			decBox := NewDecorationBox(hab, padding, margin, sentinelColor, bp)

			flb := &FillLineBox{Mode: FillEntireLine, Box: decBox}

			// Fold through layout to prove FillEntireLine allocates
			boxer := &manualBoxer{boxes: []Box{flb}}
			folder := &SimpleFolder{boxer: boxer, container: image.Rect(0, 0, 100, 100)}
			line, err := folder.Next(0)
			if err != nil {
				t.Fatalf("folder Next err: %v", err)
			}
			if line == nil {
				t.Fatalf("expected line")
			}

			img := image.NewRGBA(image.Rect(0, 0, 100, 20))
			err = line.DrawLine(img)
			if err != nil {
				t.Fatalf("DrawLine err: %v", err)
			}

			if inner.drawCalls != 1 {
				t.Fatalf("inner DrawBox calls = %d, want 1", inner.drawCalls)
			}

			if inner.drawn.Min.X != tc.expectedX {
				t.Errorf("Expected X to be %d, got %v", tc.expectedX, inner.drawn.Min.X)
			}
			if inner.drawn.Max.X != tc.expectedX+20 {
				t.Errorf("Expected Max X to be %d, got %v", tc.expectedX+20, inner.drawn.Max.X)
			}

			// Background starts at X=0 and ends at X=100 because FillEntireLine allocated 100
			// and Margin=0, Padding=10 inside the background.
			_, _, _, a1 := img.At(0, 5).RGBA()
			_, _, _, a2 := img.At(99, 5).RGBA()
			if a1 == 0 || a2 == 0 {
				t.Errorf("Background did not span allocation limits. Left Alpha: %d, Right Alpha: %d", a1, a2)
			}
		})
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
		{"Center", AlignCenter, 60},
		{"Right", AlignRight, 90},
	}

	for _, tc := range alignments {
		t.Run(tc.name, func(t *testing.T) {
			preceding := &dummyHAlignBox{width: 30, height: 10}
			inner := &dummyHAlignBox{width: 10, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}

			hab := &HorizontalAlignedBox{Box: inner, Alignment: tc.alignment}
			flb := &FillLineBox{Mode: FillRestOfLine, Box: hab}

			folder := &SimpleFolder{boxer: &manualBoxer{boxes: []Box{preceding, flb}}, container: image.Rect(0, 0, 100, 100)}

			line, err := folder.Next(0)
			if err != nil {
				t.Fatalf("folder Next err: %v", err)
			}
			if line == nil {
				t.Fatalf("expected line")
			}

			img := image.NewRGBA(image.Rect(0, 0, 100, 10))
			err = line.DrawLine(img)
			if err != nil {
				t.Fatalf("DrawLine err: %v", err)
			}

			if inner.drawn.Min.X != tc.expectedX {
				t.Errorf("Expected X=%d for FillRestOfLine %s, got %v", tc.expectedX, tc.name, inner.drawn.Min.X)
			}
		})
	}
}

func TestHorizontalAlignedBox_MultipleBoxesState(t *testing.T) {
	// Case 14: Multiple boxes retain independent state
	b1 := &HorizontalAlignedBox{Box: &dummyHAlignBox{width: 20}, Alignment: AlignLeft}
	b2 := &HorizontalAlignedBox{Box: &dummyHAlignBox{width: 20}, Alignment: AlignRight}

	if b1.Alignment == b2.Alignment {
		t.Errorf("Boxes should retain independent state")
	}

	// Put two aligned boxes on one line with separate retained allocations
	inner1 := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
	hab1 := &HorizontalAlignedBox{Box: inner1, Alignment: AlignLeft}

	inner2 := &dummyHAlignBox{width: 20, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
	hab2 := &HorizontalAlignedBox{Box: inner2, Alignment: AlignRight}

	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	dc := &DrawConfig{}

	sub1 := img.SubImage(image.Rect(0, 0, 50, 10)).(*image.RGBA)
	hab1.DrawBox(sub1, 0, dc)

	sub2 := img.SubImage(image.Rect(50, 0, 100, 10)).(*image.RGBA)
	hab2.DrawBox(sub2, 0, dc)

	if inner1.drawn.Min.X != 0 {
		t.Errorf("Left aligned box in Slot 1 should be X=0, got %d", inner1.drawn.Min.X)
	}

	if inner2.drawn.Min.X != 80 {
		t.Errorf("Right aligned box in Slot 2 should be X=80, got %d", inner2.drawn.Min.X)
	}
}

func TestHorizontalAlignedBox_ZeroWidthContent(t *testing.T) {
	// Tests that zero-width content receives an empty, correctly offset SubImage
	inner := &dummyHAlignBox{width: 0, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
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
	inner := &dummyHAlignBox{width: 100, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
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
	inner := &dummyHAlignBox{width: 120, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
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

func TestHorizontalAlignedBox_FractionalAdvanceLeftDefaultEquivalence(t *testing.T) {
	// Proves that a box with fractional width behaves exactly the same dynamically whether Default or explicitly Left Aligned
	// And checks fractional rounding semantics.
	inner1 := &dummyHAlignBox{fracWidth: fixed.I(20) + fixed.I(1)/2, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}
	inner2 := &dummyHAlignBox{fracWidth: fixed.I(20) + fixed.I(1)/2, height: 10, drawn: image.Rect(-1, -1, -1, -1), drawCalls: 0}

	hab := &HorizontalAlignedBox{Box: inner1, Alignment: AlignLeft}

	img1 := image.NewRGBA(image.Rect(0, 0, 100, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 100, 10))
	dc := &DrawConfig{}

	// Default layout passes a subImage limited by advance exactly when drawn inside SimpleLine!
	// So we simulate it directly:
	sub1 := img1.SubImage(image.Rect(0, 0, 21, 10)).(*image.RGBA)
	sub2 := img2.SubImage(image.Rect(0, 0, 21, 10)).(*image.RGBA)

	hab.DrawBox(sub1, 0, dc)
	inner2.DrawBox(sub2, 0, dc)

	if inner1.drawn.Min.X != inner2.drawn.Min.X || inner1.drawn.Max.X != inner2.drawn.Max.X {
		t.Errorf("Explicit AlignLeft fractional geometry %v did not match default un-aligned behavior %v", inner1.drawn, inner2.drawn)
	}
}
