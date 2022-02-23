package wordwrap

import (
	"github.com/google/go-cmp/cmp"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"testing"
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
			SimpleWrapper: NewSimpleWrapper("Testing this!", FontFaceForTest(t)),
			r:             SpaceFor(FontFaceForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name:          "Simple one page two line test",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFaceForTest(t)),
			r:             SpaceFor(FontFaceForTest(t), "Testing this!", "Testing this!"),
			wantPages: [][]string{
				{
					"Testing this! ",
					"Testing this!",
				},
			},
			wantErr: false,
		},
		{
			name:          "Simple two page two identical lines no extra space",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFaceForTest(t)),
			r:             SpaceFor(FontFaceForTest(t), "Testing this!"),
			wantPages: [][]string{
				{"Testing this! "},
				{"Testing this!"},
			},
			wantErr: false,
		},
		{
			name:          "Simple two page one line test but there was almost space for 2 lines",
			SimpleWrapper: NewSimpleWrapper("Testing this! Testing this!", FontFaceForTest(t)),
			r:             Shrink(SpaceFor(FontFaceForTest(t), "Testing this!"), image.Pt(0, 4)),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualPages [][]string
			for {
				got, _, err := tt.SimpleWrapper.TextToRect(tt.r)
				if err != nil {
					if (err != nil) != tt.wantErr {
						t.Errorf("TextToRect() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
				}
				if got == nil || len(got) == 0 {
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
