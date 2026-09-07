package wordwrap

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"testing"
)

type mockMetricBox struct {
	Box
	m font.Metrics
	a fixed.Int26_6
}

func (m *mockMetricBox) MetricsRect() font.Metrics {
	return m.m
}

func (m *mockMetricBox) AdvanceRect() fixed.Int26_6 {
	return m.a
}

func TestDecorationBoxMetrics(t *testing.T) {
	// A font where Height > Ascent + Descent
	innerM := font.Metrics{
		Ascent:  fixed.I(10),
		Descent: fixed.I(4),
		Height:  fixed.I(16), // 10 + 4 + 2 leading
	}
	innerA := fixed.I(20)

	mb := &mockMetricBox{
		m: innerM,
		a: innerA,
	}

	pad := fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: fixed.I(1), Y: fixed.I(2)},
		Max: fixed.Point26_6{X: fixed.I(3), Y: fixed.I(4)},
	}
	mar := fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: fixed.I(5), Y: fixed.I(6)},
		Max: fixed.Point26_6{X: fixed.I(7), Y: fixed.I(8)},
	}

	db := NewDecorationBox(mb, pad, mar, nil, BgPositioningPassThrough)

	gotM := db.MetricsRect()

	// Expected Top = pad.Min.Y (2) + mar.Min.Y (6) = 8
	// Expected Bottom = pad.Max.Y (4) + mar.Max.Y (8) = 12
	expAscent := innerM.Ascent + fixed.I(8)               // 10 + 8 = 18
	expDescent := innerM.Descent + fixed.I(12)            // 4 + 12 = 16
	expHeight := innerM.Height + fixed.I(8) + fixed.I(12) // 16 + 20 = 36

	if gotM.Ascent != expAscent {
		t.Errorf("expected Ascent %v, got %v", expAscent, gotM.Ascent)
	}
	if gotM.Descent != expDescent {
		t.Errorf("expected Descent %v, got %v", expDescent, gotM.Descent)
	}
	if gotM.Height != expHeight {
		t.Errorf("expected Height %v, got %v", expHeight, gotM.Height)
	}

	gotA := db.AdvanceRect()
	// Expected advance = 20 + pad.Min.X(1) + pad.Max.X(3) + mar.Min.X(5) + mar.Max.X(7) = 20 + 16 = 36
	expAdvance := innerA + fixed.I(16)
	if gotA != expAdvance {
		t.Errorf("expected Advance %v, got %v", expAdvance, gotA)
	}
}

func TestDecorationBoxAdvance(t *testing.T) {
	innerA := fixed.I(20)
	mb := &mockMetricBox{
		m: font.Metrics{Ascent: fixed.I(10), Descent: fixed.I(4), Height: fixed.I(14)},
		a: innerA,
	}

	// Test padding only
	pad := fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: fixed.I(2)},
		Max: fixed.Point26_6{X: fixed.I(3)},
	}
	dbPad := NewDecorationBox(mb, pad, fixed.Rectangle26_6{}, nil, BgPositioningPassThrough)
	if got := dbPad.AdvanceRect(); got != innerA+fixed.I(5) {
		t.Errorf("AdvanceRect with padding = %v, want %v", got, innerA+fixed.I(5))
	}

	// Test margin only
	mar := fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: fixed.I(4)},
		Max: fixed.Point26_6{X: fixed.I(1)},
	}
	dbMar := NewDecorationBox(mb, fixed.Rectangle26_6{}, mar, nil, BgPositioningPassThrough)
	if got := dbMar.AdvanceRect(); got != innerA+fixed.I(5) {
		t.Errorf("AdvanceRect with margin = %v, want %v", got, innerA+fixed.I(5))
	}
}

func TestDecorationBoxMetricsZero(t *testing.T) {
	innerM := font.Metrics{
		Ascent:  fixed.I(10),
		Descent: fixed.I(4),
		Height:  fixed.I(16),
	}
	innerA := fixed.I(20)

	mb := &mockMetricBox{
		m: innerM,
		a: innerA,
	}

	db := NewDecorationBox(mb, fixed.Rectangle26_6{}, fixed.Rectangle26_6{}, nil, BgPositioningPassThrough)

	gotM := db.MetricsRect()
	if gotM.Ascent != innerM.Ascent || gotM.Descent != innerM.Descent || gotM.Height != innerM.Height {
		t.Errorf("MetricsRect with zero decoration = %v, want %v", gotM, innerM)
	}

	gotA := db.AdvanceRect()
	if gotA != innerA {
		t.Errorf("AdvanceRect with zero decoration = %v, want %v", gotA, innerA)
	}
}

func TestDecorationBoxNested(t *testing.T) {
	innerM := font.Metrics{Ascent: fixed.I(10), Descent: fixed.I(4), Height: fixed.I(14)}
	innerA := fixed.I(20)
	mb := &mockMetricBox{m: innerM, a: innerA}

	pad := fixed.Rectangle26_6{Min: fixed.Point26_6{X: fixed.I(1), Y: fixed.I(1)}, Max: fixed.Point26_6{X: fixed.I(1), Y: fixed.I(1)}}

	db1 := NewDecorationBox(mb, pad, pad, nil, BgPositioningPassThrough)  // Total add: 2+2=4 padding+margin per dimension
	db2 := NewDecorationBox(db1, pad, pad, nil, BgPositioningPassThrough) // Another 4

	gotM := db2.MetricsRect()
	if gotM.Height != innerM.Height+fixed.I(8) { // 14 + 8 = 22
		t.Errorf("Nested Height = %v, want %v", gotM.Height, innerM.Height+fixed.I(8))
	}

	gotA := db2.AdvanceRect()
	if gotA != innerA+fixed.I(8) {
		t.Errorf("Nested Advance = %v, want %v", gotA, innerA+fixed.I(8))
	}
}
