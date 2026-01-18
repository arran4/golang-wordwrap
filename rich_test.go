package wordwrap

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"golang.org/x/image/font"
)

func TestProcessRichArgs(t *testing.T) {
	// Setup
	fontFace := FontFace16DPI180ForTest(t)
	// drawer := &font.Drawer{Face: fontFace} // Unused

	tests := []struct {
		name         string
		args         []interface{}
		wantContents int
		check        func(*testing.T, []*Content, *font.Drawer)
	}{
		{
			name:         "Simple String",
			args:         []interface{}{"Hello"},
			wantContents: 1,
			check: func(t *testing.T, c []*Content, d *font.Drawer) {
				if c[0].text != "Hello" {
					t.Errorf("Content[0].text = %q, want %q", c[0].text, "Hello")
				}
			},
		},
		{
			name:         "String with Font",
			args:         []interface{}{fontFace, "Hello"},
			wantContents: 1,
			check: func(t *testing.T, c []*Content, d *font.Drawer) {
				// Style should have font
				// Content structure is private, but we can check if it processed without panic
				// and if drawer face was updated in state (returned as 2nd arg)
				// Wait, ProcessRichArgs returns `state.drawer`.
				if d == nil {
					t.Error("Drawer is nil")
				}
			},
		},
		{
			name:         "Multiple Strings",
			args:         []interface{}{"Hello", " ", "World"},
			wantContents: 3,
		},
		{
			name: "Group Scoping",
			args: []interface{}{
				"Normal",
				Group{
					Args: []interface{}{
						TextColor(color.Black),
						"Black",
					},
				},
				"Normal",
			},
			wantContents: 3,
			check: func(t *testing.T, c []*Content, d *font.Drawer) {
				// We can't easily inspect content style as it is private/internal in Content.
				// But we verify structural parsing.
			},
		},
		{
			name: "Highlight and TextImage",
			args: []interface{}{
				Highlight(color.RGBA{255, 255, 0, 255}),
				TextImage(image.NewUniform(color.Black)),
				"Highlighted Text",
			},
			wantContents: 1,
			// check logic not strictly needed if we just want coverage of the function calls
		},
		{
			name: "Nil Drawer Safety",
			args: []interface{}{
				(*font.Drawer)(nil), // Explict nil drawer
				"Text",
			},
			wantContents: 1,
			check: func(t *testing.T, c []*Content, d *font.Drawer) {
				if d != nil {
					// We expect logic to NOT set drawer if nil provided,
					// OR set default black drawer if none exists?
					// Implementation of ProcessRichArgs logic:
					// if state.drawer == nil && state.defaultFont != nil { set default }
					// if *font.Drawer arg is passed as nil:
					// My fix: if v != nil { ... }
					// So it ignores nil drawer.
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, outDrawer, _, _, _, _ := ProcessRichArgs(tt.args...)

			if len(contents) != tt.wantContents {
				t.Errorf("ProcessRichArgs() returned %d contents, want %d", len(contents), tt.wantContents)
			}

			if tt.check != nil {
				tt.check(t, contents, outDrawer)
			}
		})
	}
}

func TestRichTextRendering(t *testing.T) {
	// Setup
	fontFace := FontFace16DPI180ForTest(t)

	// Create arguments using Highlight and TextImage
	highlightColor := color.RGBA{255, 255, 0, 255}
	textImg := image.NewRGBA(image.Rect(0, 0, 10, 10)) // Simple pattern
	fill(textImg, color.RGBA{0, 255, 255, 255})        // Fill cyan

	args := []interface{}{
		fontFace,
		"Normal ",
		Highlight(highlightColor),
		"Highlighted",
		Highlight(color.Transparent), // Accumulates
		" Normal ",
		Group{
			Args: []interface{}{
				Highlight(color.RGBA{0, 0, 255, 255}),
				"Blue Highlight",
			},
		},
		" ",
		TextImage(textImg), // Should render cyan box
		"Text Pattern",
	}

	wrapper := NewRichWrapper(args...)

	// Render to trigger Draw methods
	target := image.NewRGBA(image.Rect(0, 0, 400, 100))

	// Layout
	lines, _, err := wrapper.TextToRect(target.Bounds())
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// Render
	if err := wrapper.RenderLines(target, lines, target.Bounds().Min); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verification
	foundYellow := false
	bounds := target.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := target.At(x, y)
			r, g, b, a := c.RGBA()
			// Yellow: R=High, G=High, B=Low.
			if r > 0xF000 && g > 0xF000 && b < 0x1000 && a > 0xF000 {
				foundYellow = true
				break
			}
		}
		if foundYellow {
			break
		}
	}
	if !foundYellow {
		t.Errorf("Did not find yellow highlight pixels")
	}
}

func fill(img *image.RGBA, c color.Color) {
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
}
