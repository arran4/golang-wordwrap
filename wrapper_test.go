package wordwrap

import (
	_ "embed"
	"image"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func TestSimpleWrapper_TextToRect(t *testing.T) {
	tests := []struct {
		name          string
		SimpleWrapper *SimpleWrapper
		r             image.Rectangle
		wantPages     [][]string
		wantErr       bool
	}{
		{
			name:          "Simple one page one line test",
			SimpleWrapper: NewSimpleWrapper("Testing this!", FontFace16DPI180ForTest(t)),
			r:             SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name:          "Simple one page two line test",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t)),
			r:             SpaceFor(FontFace16DPI180ForTest(t), "Testing this!", "Testing this!"),
			wantPages: [][]string{
				{
					"Testing this!  ",
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name:          "Simple two page two identical lines no extra space",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t)),
			r:             SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this!  "},
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name:          "Simple two page one line test but there was almost space for 2 lines",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t)),
			r:             Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this!  ",
				},
				{
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Simple two page one line test with a continue box on the first one but otherwise would be " +
				"the first word of the 2nd line. Unicode",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t),
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵"))),
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this! ↵",
				},
				{
					"Testing  this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Page break too big",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t),
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵↵↵↵↵↵↵↵↵↵↵"))),
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
			wantErr: true,
		},
		{
			name: "Page Break two page one line test, page break pushes 2nd line into next box",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t),
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵"))),
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! ↵↵", "Testing this! ↵↵"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this! ↵",
				},
				{
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Page Break two page one line test, page break pushes 2nd line into next box white space is not folded",
			SimpleWrapper: NewSimpleWrapper("Testing this!     Testing this!", FontFace16DPI180ForTest(t),
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵"))),
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! ↵↵", "Testing this! ↵↵"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this!     ↵",
				},
				{
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Page break doesn't fit in rect",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFace16DPI180ForTest(t),
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵"))),
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! ↵↵↵"), image.Pt(0, 4)),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualPages [][]string
			for {
				_ = tt.SimpleWrapper.HasNext()
				// Check consistency if desired, but mostly just exercising the method
				got, _, err := tt.SimpleWrapper.TextToRect(tt.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("TextToRect() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) == 0 {
					break
				}
				var actualLines []string
				for _, line := range got {
					actualLines = append(actualLines, line.TextValue())
				}
				actualPages = append(actualPages, actualLines)
			}
			if s := cmp.Diff(tt.wantPages, actualPages); s != "" {
				t.Errorf("TextToRect(): \n %s", s)
			}
		})
	}
}

func NewSimpleTextBoxForTest(t *testing.T, ff font.Face, s string) Box {
	d := &font.Drawer{
		Face: ff,
	}
	b, err := NewSimpleTextBox(d, s)
	if err != nil {
		t.Fatalf("Error creating box: %s", err)
	}
	return b
}

func Shrink(spaceFor image.Rectangle, pt image.Point) image.Rectangle {
	spaceFor.Max = spaceFor.Max.Sub(pt)
	return spaceFor
}

func SpaceFor(ff font.Face, s ...string) image.Rectangle {
	d := font.Drawer{
		Face: ff,
	}
	var a = fixed.I(0)
	var h fixed.Int26_6
	for _, l := range s {
		_, aa := d.BoundString(l)
		if aa > a {
			a = aa
		}
		m := ff.Metrics()
		h = h + m.Descent + m.Ascent
	}
	return image.Rect(0, 0, a.Ceil()+2, h.Ceil())
}

func TestSimpleWrapTextToImage(t *testing.T) {
	// Simple sanity check test
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	err := SimpleWrapTextToImage("Hello World", img, FontFace16DPI180ForTest(t))
	if err != nil {
		t.Errorf("SimpleWrapTextToImage failed: %v", err)
	}
	// We can't easily verify pixels, but we verify it runs without error.
}

func TestSimpleWrapTextToRect_Helper(t *testing.T) {
	sw, lines, p, err := SimpleWrapTextToRect("Hello World", image.Rect(0, 0, 100, 100), FontFace16DPI180ForTest(t))
	if err != nil {
		t.Errorf("SimpleWrapTextToRect failed: %v", err)
	}
	if sw == nil {
		t.Error("Returned wrapper is nil")
	}
	if len(lines) == 0 {
		t.Error("No lines returned")
	}
	if p == (image.Point{}) {
		t.Error("Point is zero (might be valid but unlikely for non-empty text)")
	}
}

func TestBlockAlignment(t *testing.T) {
	// Test Vertical and Horizontal block alignment logic by inspecting offset of rendered lines
	// Since RenderLines calculates offset internally, we might need to inspect it or trust it works.
	// But `calculateAlignmentOffset` is a method on SimpleWrapper.

	sw := NewSimpleWrapper("", FontFace16DPI180ForTest(t))

	// Create mock lines of known size
	lines := []Line{
		&SimpleLine{size: fixed.R(0, 0, 100, 20)}, // 100x20
		&SimpleLine{size: fixed.R(0, 0, 50, 20)},  // 50x20
	}
	// Total block size: 100 x 40.

	bounds := image.Rect(0, 0, 200, 200)

	// Test Default (Top Left)
	off := sw.calculateAlignmentOffset(lines, bounds)
	if off != (image.Point{0, 0}) {
		t.Errorf("Default offset = %v, want (0,0)", off)
	}

	// Test Center/Center
	sw.SetHorizontalBlockPosition(HorizontalCenterBlock)
	sw.SetVerticalBlockPosition(VerticalCenterBlock)
	off = sw.calculateAlignmentOffset(lines, bounds)
	// X: (200 - 100)/2 = 50
	// Y: (200 - 40)/2 = 80
	want := image.Point{50, 80}
	if off != want {
		t.Errorf("Center/Center offset = %v, want %v", off, want)
	}

	// Test Right/Bottom
	sw.SetHorizontalBlockPosition(RightBlock)
	sw.SetVerticalBlockPosition(BottomBlock)
	off = sw.calculateAlignmentOffset(lines, bounds)
	// X: 200 - 100 = 100
	// Y: 200 - 40 = 160
	want = image.Point{100, 160}
	if off != want {
		t.Errorf("Right/Bottom offset = %v, want %v", off, want)
	}
}
