package wordwrap

import (
	"image"
	"testing"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
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
		name          string
		alignment     HorizontalAlignment
		boxWidth      int
		allocWidth    int
		expectedMinX  int
	}{
		{"AlignLeft", AlignLeft, 20, 100, 0},
		{"AlignCenter", AlignCenter, 20, 100, 40},
		{"AlignRight", AlignRight, 20, 100, 80},
		{"AlignRight_Overflow", AlignRight, 120, 100, 0}, // No offset if natural > allocated
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

			img := image.NewRGBA(image.Rect(0, 0, tc.allocWidth, 10))
			dc := &DrawConfig{}
			hab.DrawBox(img, 0, dc)

			if inner.drawn.Min.X != tc.expectedMinX {
				t.Errorf("Expected Min.X = %d, got %d", tc.expectedMinX, inner.drawn.Min.X)
			}
		})
	}
}
