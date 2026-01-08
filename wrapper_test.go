package wordwrap

import (
	"bytes"
	_ "embed"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/png"
	"testing"
)

var (
	//go:embed "testdata/chevron.png"
	chevronImageBytes []byte
)

func chevronImage() Image {
	i, _ := png.Decode(bytes.NewReader(chevronImageBytes))
	return i.(Image)
}

func TestSimpleWrapper_TextToRect(t *testing.T) {
	tests := []struct {
		name      string
		opts      []WrapperOption
		contents  []*Content
		r         image.Rectangle
		wantPages [][]string
		wantErr   bool
	}{
		{
			name: "Simple one page one line test",
			contents: []*Content{
				NewContent("Testing this!"),
			},
			r: SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name: "Simple one page two line test",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI180ForTest(t), "Testing this!", "Testing this!"),
			wantPages: [][]string{
				{
					"Testing this! ",
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Simple two page two identical lines no extra space",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this! "},
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name: "Simple two page one line test but there was almost space for 2 lines",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this! ",
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
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
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
			name: "Page break too big",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵↵↵↵↵↵↵↵↵↵↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
			wantErr: true,
		},
		{
			name: "Page Break two page one line test, page break pushes 2nd line into next box",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
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
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this!     Testing this!"),
			},
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
			name: "Page Break two page one line test, page break pushes 2nd line into next box - image",
			opts: []WrapperOption{
				NewPageBreakBox(NewImageBox(chevronImage())),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI75ForTest(t), "Testing this! ↵↵ ↵↵ ↵↵", "Testing this! ↵↵ ↵↵ ↵↵"),
			wantPages: [][]string{
				{
					"Testing this! ",
				},
				{
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Page break images can cause failures too if too tall",
			opts: []WrapperOption{
				NewPageBreakBox(NewImageBox(chevronImage())),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       SpaceFor(FontFace16DPI75ForTest(t), "Testing this! ↵↵ ↵↵ ↵↵"),
			wantErr: true,
		},
		{
			name: "Page break doesn't fit in rect",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! ↵↵↵"), image.Pt(0, 4)),
			wantErr: true,
		},
		{
			name: "Simple one page two line test",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI180ForTest(t), "Testing this!", "Testing this!"),
			wantPages: [][]string{
				{
					"Testing this! ",
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Simple two page two identical lines no extra space",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this! "},
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name: "Simple two page one line test but there was almost space for 2 lines",
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this!"), image.Pt(0, 4)),
			wantPages: [][]string{
				{
					"Testing this! ",
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
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
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
			name: "Page break too big",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace16DPI180ForTest(t), "↵↵↵↵↵↵↵↵↵↵↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! Testing"), image.Pt(0, 4)),
			wantErr: true,
		},
		{
			name: "Page Break two page one line test, page break pushes 2nd line into next box",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
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
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this!     Testing this!"),
			},
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
			name: "Page Break two page one line test, page break pushes 2nd line into next box - image",
			opts: []WrapperOption{
				NewPageBreakBox(NewImageBox(chevronImage())),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r: SpaceFor(FontFace16DPI75ForTest(t), "Testing this! ↵↵ ↵↵ ↵↵", "Testing this! ↵↵ ↵↵ ↵↵"),
			wantPages: [][]string{
				{
					"Testing this! ",
				},
				{
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name: "Page break images can cause failures too if too tall",
			opts: []WrapperOption{
				NewPageBreakBox(NewImageBox(chevronImage())),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       SpaceFor(FontFace16DPI75ForTest(t), "Testing this! ↵↵ ↵↵ ↵↵"),
			wantErr: true,
		},
		{
			name: "Page break doesn't fit in rect",
			opts: []WrapperOption{
				NewPageBreakBox(NewSimpleTextBoxForTest(t, FontFace24DPI180ForTest(t), "↵")),
			},
			contents: []*Content{
				NewContent("Testing this! Testing this!"),
			},
			r:       Shrink(SpaceFor(FontFace16DPI180ForTest(t), "Testing this! ↵↵↵"), image.Pt(0, 4)),
			wantErr: true,
		},
		{
			name: "Styled content",
			contents: []*Content{
				NewContent("Hello, ", WithFont(FontFace16DPI180ForTest(t))),
				NewContent("world!", WithFont(FontFace24DPI180ForTest(t))),
			},
			r: SpaceFor(FontFace24DPI180ForTest(t), "Hello, world!"),
			wantPages: [][]string{
				{"Hello, world!"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSimpleWrapper(FontFace16DPI180ForTest(t), tt.opts, tt.contents...)
			var actualPages [][]string
			var boxes []Box
			for {
				got, _, err := sw.TextToRect(tt.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("TextToRect() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) == 0 {
					if !sw.HasNext() {
						break
					}
				}
				var actualLines []string
				for _, line := range got {
					actualLines = append(actualLines, line.TextValue())
					boxes = append(boxes, line.Boxes()...)
				}
				actualPages = append(actualPages, actualLines)
			}
			if s := cmp.Diff(tt.wantPages, actualPages); s != "" {
				t.Errorf("TextToRect(): \n %s", s)
			}
			if tt.name == "Styled content" {
				if boxes[0].FontDrawer().Face != FontFace16DPI180ForTest(t) {
					t.Errorf("expected font face %+v, got %+v", FontFace16DPI180ForTest(t), boxes[0].FontDrawer().Face)
				}
				if boxes[len(boxes)-1].FontDrawer().Face != FontFace24DPI180ForTest(t) {
					t.Errorf("expected font face %+v, got %+v", FontFace24DPI180ForTest(t), boxes[len(boxes)-1].FontDrawer().Face)
				}
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
