package wordwrap

import (
	"image"
	"image/color"
	"testing"
)

func TestSimpleWrapper_TextToSpecs(t *testing.T) {
	font := FontFace16DPI180ForTest(t)
	text := "Hello World"
	// Measure text size approximately
	// "Hello World" at 16pt 180dpi.
	// We can let the test dynamic measure it first to verify constraints relative to it.

	wrapper := NewSimpleWrapper([]*Content{{text: text}}, font)

	// Measure natural size
	lines, naturalSize, err := wrapper.TextToRect(image.Rect(0, 0, 10000, 10000))
	if err != nil {
		t.Fatalf("Failed to measure natural size: %v", err)
	}

	naturalWidth := 0
	for _, l := range lines {
		if w := l.Size().Dx(); w > naturalWidth {
			naturalWidth = w
		}
	}
	naturalHeight := naturalSize.Y

	tests := []struct {
		name             string
		opts             []SpecOption
		wantPageSize     image.Point
		wantContentStart image.Point
		wantMargin       SpecMargin
		checkConstraint  func(t *testing.T, natural, result image.Point)
	}{
		{
			name:             "Default Unbounded",
			opts:             []SpecOption{},
			wantPageSize:     image.Point{X: naturalWidth, Y: naturalHeight},
			wantContentStart: image.Point{0, 0},
		},
		{
			name:         "Fixed Width Larger",
			opts:         []SpecOption{Width(Fixed(naturalWidth + 100))},
			wantPageSize: image.Point{X: naturalWidth + 100, Y: naturalHeight},
		},
		{
			name:         "Fixed Height Larger",
			opts:         []SpecOption{Height(Fixed(naturalHeight + 100))},
			wantPageSize: image.Point{X: naturalWidth, Y: naturalHeight + 100},
		},
		{
			name: "Min Width (Active)",
			opts: []SpecOption{Width(Min(Fixed(naturalWidth+50), Unbounded()))},
			// Min(Fixed(Large), Unbounded) -> Min(Large, Natural) -> Natural
			wantPageSize: image.Point{X: naturalWidth, Y: naturalHeight},
		},
		{
			name:         "Max Width (Active)", // Acts as "At least"
			opts:         []SpecOption{Width(Max(Fixed(naturalWidth+50), Unbounded()))},
			wantPageSize: image.Point{X: naturalWidth + 50, Y: naturalHeight},
		},
		{
			name: "Small Width (Wrapping)",
			opts: []SpecOption{Width(Fixed(naturalWidth / 2))},
			checkConstraint: func(t *testing.T, n, r image.Point) {
				if r.X != naturalWidth/2 {
					t.Errorf("Expected width %d, got %d", naturalWidth/2, r.X)
				}
				if r.Y <= n.Y {
					t.Errorf("Expected height to grow due to wrapping, got %d <= %d", r.Y, n.Y)
				}
			},
		},
		{
			name:             "Margins",
			opts:             []SpecOption{Padding(10, color.Black)},
			wantContentStart: image.Point{10, 10},
			wantMargin:       SpecMargin{10, 10, 10, 10, color.Black},
			checkConstraint: func(t *testing.T, n, r image.Point) {
				// Page size should be Natural + Margins (if unbounded)
				// Content is laid out in unbounded.
				// Width = Natural + Left + Right
				// Height = Natural + Top + Bottom
				expectedW := naturalWidth + 20
				expectedH := naturalHeight + 20
				if r.X != expectedW {
					t.Errorf("Expected width %d, got %d", expectedW, r.X)
				}
				if r.Y != expectedH {
					t.Errorf("Expected height %d, got %d", expectedH, r.Y)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Re-create wrapper to allow Reset tests implicitly (TextToSpecs uses Reset)
			// But wrapper is stateful if used? TextToSpecs calls Reset at start.
			// Let's use the shared wrapper to verify Reset works!

			res, err := wrapper.TextToSpecs(tt.opts...)
			if err != nil {
				t.Fatalf("TextToSpecs failed: %v", err)
			}

			if tt.checkConstraint != nil {
				tt.checkConstraint(t, image.Point{naturalWidth, naturalHeight}, res.PageSize)
			} else {
				if res.PageSize.X != tt.wantPageSize.X && tt.wantPageSize.X != 0 {
					t.Errorf("PageSize.X = %d, want %d", res.PageSize.X, tt.wantPageSize.X)
				}
				if res.PageSize.Y != tt.wantPageSize.Y && tt.wantPageSize.Y != 0 {
					t.Errorf("PageSize.Y = %d, want %d", res.PageSize.Y, tt.wantPageSize.Y)
				}
			}

			if res.ContentStart != tt.wantContentStart {
				t.Errorf("ContentStart = %v, want %v", res.ContentStart, tt.wantContentStart)
			}
			if tt.wantMargin != (SpecMargin{}) {
				if res.Margin != tt.wantMargin {
					t.Errorf("Margin = %v, want %v", res.Margin, tt.wantMargin)
				}
			}
		})
	}
}

func TestSizeFunctions(t *testing.T) {
	n := 100

	if val := Unbounded()(n); val != n {
		t.Errorf("Unbounded(%d) = %d, want %d", n, val, n)
	}

	if val := Fixed(50)(n); val != 50 {
		t.Errorf("Fixed(50)(%d) = %d, want 50", n, val)
	}

	if val := Min(Fixed(50), Unbounded())(n); val != 50 {
		t.Errorf("Min(50, 100) = %d, want 50", val)
	}

	if val := Max(Fixed(50), Unbounded())(n); val != 100 {
		t.Errorf("Max(50, 100) = %d, want 100", val)
	}

	if val := Max(Fixed(150), Unbounded())(n); val != 150 {
		t.Errorf("Max(150, 100) = %d, want 150", val)
	}

	// DPI check: 96 DPI. 1 inch = 96px.
	// DPI(96)(72) => 96 (since input is points, 72pts = 1 inch)
	if val := DPI(96)(72); val != 96 {
		t.Errorf("DPI(96)(72) = %d, want 96", val)
	}
}
