package wordwrap

import (
	"image"
	"image/color"
	"reflect"
	"testing"

	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func TestSimpleBoxer_BoxNextWord(t *testing.T) {
	grf := FontFace16DPI180ForTest(t)
	type args struct {
		fce   font.Face
		color image.Image
		text  string
	}
	tests := []struct {
		name             string
		args             args
		wantBoxString    string
		wantSimpleBox    bool
		wantLineBreakBox bool
		wantN            int
		wantErr          error
		wantNilBox       bool
	}{
		{
			name: "One word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "OnEWorD",
			},
			wantBoxString: "OnEWorD",
			wantN:         len("OnEWorD"),
			wantSimpleBox: true,
		},
		{
			name: "Empty string",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "",
			},
			wantBoxString: "",
			wantN:         len(""),
			wantSimpleBox: true,
		},
		{
			name: "Multiple spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "   ",
			},
			wantBoxString: "   ",
			wantN:         len("   "),
			wantSimpleBox: true,
		},
		{
			name: "Two words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "two words",
			},
			wantBoxString: "two",
			wantN:         len("two"),
			wantSimpleBox: true,
		},
		{
			name: "Two words multiple spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "two    words",
			},
			wantBoxString: "two",
			wantN:         len("two"),
			wantSimpleBox: true,
		},
		{
			name: "multiple spaces then one word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "    words",
			},
			wantBoxString: "    ",
			wantN:         len("    "),
			wantSimpleBox: true,
		},
		{
			name: "Line break CRLF breaks words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "words\r\nhello",
			},
			wantBoxString: "words",
			wantN:         len("words"),
			wantSimpleBox: true,
		},
		{
			name: "Line break LF breaks words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "words\nhello",
			},
			wantBoxString: "words",
			wantN:         len("words"),
			wantSimpleBox: true,
		},
		{
			name: "Line break CRLF breaks spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "    \r\nhello",
			},
			wantBoxString: "    ",
			wantN:         len("    "),
			wantSimpleBox: true,
		},
		{
			name: "Line break LF breaks spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "    \nhello",
			},
			wantBoxString: "    ",
			wantN:         len("    "),
			wantSimpleBox: true,
		},
		{
			name: "Captures LF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\n",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\nhello",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\n ",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\nword",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n    ",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFCRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n\n\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Empty returns nil",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "",
			},
			wantNilBox: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewSimpleBoxer([]*Content{{text: tt.args.text}}, &font.Drawer{
				Src:  tt.args.color,
				Face: tt.args.fce,
			})
			b, n, err := sb.Next()
			if tt.wantSimpleBox {
				sb, ok := b.(*SimpleTextBox)
				if ok {
					if !reflect.DeepEqual(sb.Contents, tt.wantBoxString) {
						t.Errorf("BoxNextWord()[0].Contents b = %v, wantBoxString %v", sb.Contents, tt.wantBoxString)
					}
				} else {
					if len(tt.wantBoxString) > 0 {
						t.Errorf("BoxNextWord()[0].Contents b = %v, wantBoxString %v", b, tt.wantBoxString)
					}
				}
			} else if tt.wantNilBox {
				if b != nil {
					t.Errorf("BoxNextWord()[0] b = %s, wanted nil", reflect.TypeOf(b))
				}
			} else if tt.wantLineBreakBox {
				_, ok := b.(*LineBreakBox)
				if !ok {
					t.Errorf("BoxNextWord()[0] b = %s, wanted line break", reflect.TypeOf(b))
				}
			} else {
				t.Errorf("Unselected want for BoxNextWord()[0]")
			}
			if n != tt.wantN {
				t.Errorf("BoxNextWord() n = %v, wantN %v", n, tt.wantN)
			}
			if err != tt.wantErr {
				t.Errorf("BoxNextWord().error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func FontFace16DPI180ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 180, gr)
	return grf
}

func FontFace16DPI75ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 75, gr)
	return grf
}

func FontFace24DPI180ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(24, 180, gr)
	return grf
}

func TestSimpleBoxer_Reset(t *testing.T) {
	text := "Hello World"
	fontFace := FontFace16DPI180ForTest(t)
	drawer := &font.Drawer{Face: fontFace, Src: image.NewUniform(color.Black)}

	contents := []*Content{{text: text}}
	boxer := NewSimpleBoxer(contents, drawer)

	// Consume some boxes
	_, _, err := boxer.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	// Reset
	boxer.Reset()

	// Should be at start
	// "Hello" (SimpleTextBox)
	b, _, err := boxer.Next()
	if err != nil {
		t.Fatalf("Next after Reset failed: %v", err)
	}

	sb, ok := b.(*SimpleTextBox)
	if !ok {
		t.Fatalf("Expected SimpleTextBox, got %T", b)
	}

	// Depending on boxer logic, first box of "Hello World" is "Hello"
	if sb.Contents != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", sb.Contents)
	}
}

func TestImageBoxMetricsInSimpleBoxer(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	content := NewImageContent(img)

	boxer := NewSimpleBoxer([]*Content{content}, nil)
	b, _, err := boxer.Next()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ib, ok := b.(*ImageBox)
	if !ok {
		t.Fatalf("Expected *ImageBox, got %T", b)
	}

	m := ib.MetricsRect()
	if m.Ascent.Ceil() != 30 {
		t.Errorf("Expected ascent to be 30, got %d", m.Ascent.Ceil())
	}
	if m.Height.Ceil() != 30 {
		t.Errorf("Expected height to be 30, got %d", m.Height.Ceil())
	}
}

// Minimal mock for testing BackgroundBox DrawBox bounds logic
type mockMetricsBox struct {
	dummyBox // embed to inherit other methods if any exist in the same package
	ascent   int
	descent  int
	drawnY   fixed.Int26_6
	drawnDc  *DrawConfig
}

func (m *mockMetricsBox) AdvanceRect() fixed.Int26_6 {
	return fixed.I(20)
}

func (m *mockMetricsBox) MetricsRect() font.Metrics {
	return font.Metrics{
		Ascent:  fixed.I(m.ascent),
		Descent: fixed.I(m.descent),
	}
}

func (m *mockMetricsBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	m.drawnY = y
	m.drawnDc = dc
}

func (m *mockMetricsBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	return fixed.I(20), fixed.I(m.ascent + m.descent)
}

func (m *mockMetricsBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	return fixed.I(20), fixed.I(m.ascent + m.descent)
}

func TestBackgroundBox_DrawBox_BoundsClipsToMetrics(t *testing.T) {
	inner := &mockMetricsBox{ascent: 15, descent: 5}

	// Create a deterministic pattern image (gradient/stripes)
	// so we can verify PassThrough sampling didn't shift
	bgSrc := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			bgSrc.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}

	bb := &BackgroundBox{
		Box:           inner,
		Background:    bgSrc,
		BgPositioning: BgPositioningPassThrough,
	}

	// The line provides a full canvas, taller than the inner box
	// We'll create it as a sub-image to verify bounds.Min shifting works correctly too
	fullImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	img := fullImg.SubImage(image.Rect(10, 20, 110, 70)).(Image) // Bounds: X:10->110, Y:20->70

	dc := &DrawConfig{}

	// Draw at baseline Y=30 (relative to bounds.Min.Y=20, so drawn Y=50)
	y := fixed.I(30)

	bb.DrawBox(img, y, dc)

	if inner.drawnY != y {
		t.Errorf("Expected inner box to be drawn at Y=%v, got %v", y, inner.drawnY)
	}

	// Expected drawn region is Min.Y + (30 - 15) = 20 + 15 = 35
	// to Min.Y + (30 + 5) = 20 + 35 = 55
	// Check a pixel outside this region (e.g., Y=34 and Y=56)
	cTop := fullImg.RGBAAt(20, 34)
	if cTop.A != 0 {
		t.Errorf("Expected pixel above bounds (Y=34) to remain empty, got %v", cTop)
	}

	cBottom := fullImg.RGBAAt(20, 56)
	if cBottom.A != 0 {
		t.Errorf("Expected pixel below bounds (Y=56) to remain empty, got %v", cBottom)
	}

	// Check a pixel inside this region (e.g., Y=40, X=20)
	cInside := fullImg.RGBAAt(20, 40)
	if cInside.A != 255 {
		t.Errorf("Expected pixel inside bounds (X=20, Y=40) to be painted, got %v", cInside)
	}

	// Verify PassThrough semantics: Source coordinate should match destination coordinate
	// For PassThrough, if we sample destination X=20, Y=40, it should correspond to bgSrc X=20, Y=40
	expectedColor := color.RGBA{uint8(20), uint8(40), 0, 255}
	if cInside != expectedColor {
		t.Errorf("PassThrough sampling shifted! Expected %v, got %v at X=20, Y=40", expectedColor, cInside)
	}
}
