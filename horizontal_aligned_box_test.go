package wordwrap

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"testing"
)

type dummyHAlignBox struct {
	width  int
	height int
	drawn  image.Rectangle
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
