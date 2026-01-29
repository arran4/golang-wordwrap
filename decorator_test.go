package wordwrap

import (
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// MockBox implements Box for testing
type MockBox struct {
	width, height, ascent, descent fixed.Int26_6
}

func (mb *MockBox) AdvanceRect() fixed.Int26_6 { return mb.width }
func (mb *MockBox) MetricsRect() font.Metrics {
	return font.Metrics{Ascent: mb.ascent, Descent: mb.descent, Height: mb.height}
}
func (mb *MockBox) Whitespace() bool                                      { return false }
func (mb *MockBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig)      {}
func (mb *MockBox) FontDrawer() *font.Drawer                              { return nil }
func (mb *MockBox) Len() int                                              { return 0 }
func (mb *MockBox) TextValue() string                                     { return "" }
func (mb *MockBox) MinSize() (fixed.Int26_6, fixed.Int26_6)               { return mb.width, mb.height }
func (mb *MockBox) MaxSize() (fixed.Int26_6, fixed.Int26_6)               { return mb.width, mb.height }

func TestDecorationBoxMetrics(t *testing.T) {
	inner := &MockBox{
		width:   fixed.I(10),
		height:  fixed.I(10),
		ascent:  fixed.I(8),
		descent: fixed.I(2),
	}

	pad := fixed.Rectangle26_6{
		Min: fixed.P(2, 3), // Left: 2, Top: 3
		Max: fixed.P(4, 5), // Right: 4, Bottom: 5
	}
	// Margin 0
	margin := fixed.Rectangle26_6{}

	db := NewDecorationBox(inner, pad, margin, nil, BgPositioningZeroed)

	// Check AdvanceRect
	expectedWidth := inner.width + pad.Min.X + pad.Max.X
	if got := db.AdvanceRect(); got != expectedWidth {
		t.Errorf("AdvanceRect: expected %d, got %d", expectedWidth, got)
	}

	// Check MetricsRect
	m := db.MetricsRect()
	expectedAscent := inner.ascent + pad.Min.Y
	expectedDescent := inner.descent + pad.Max.Y

	if m.Ascent != expectedAscent {
		t.Errorf("Ascent: expected %d, got %d", expectedAscent, m.Ascent)
	}
	if m.Descent != expectedDescent {
		t.Errorf("Descent: expected %d, got %d", expectedDescent, m.Descent)
	}
}
