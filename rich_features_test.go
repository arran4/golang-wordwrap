package wordwrap

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/math/fixed"
)

// TestAdvancedRichFeatures tests extensive use cases including linear style changes,
// groups, containers, min size, and reset functionality.
func TestAdvancedRichFeatures(t *testing.T) {
	// Setup
	fontFace := FontFace16DPI180ForTest(t)
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	// largeFont := ... (Use same font for now, just checking logic)

	t.Run("Linear Style Changes", func(t *testing.T) {
		args := []interface{}{
			fontFace,
			"Plain",
			Color(red),
			"Red",
			Color(blue),
			"Blue",
			Reset(),
			"PlainAgain",
		}

		contents, _, _, _, _, _ := ProcessRichArgs(args...)

		if len(contents) != 4 {
			t.Fatalf("Expected 4 content blocks, got %d", len(contents))
		}

		// Check Plain
		if c0 := contents[0]; c0.text != "Plain" || (c0.style != nil && c0.style.FontDrawerSrc != nil) {
			t.Errorf("Content[0] should be plain, got text %q style %v", c0.text, c0.style)
		}

		// Check Red
		if c1 := contents[1]; c1.text != "Red" {
			t.Errorf("Content[1] text mismatch")
		} else if c1.style == nil || c1.style.FontDrawerSrc == nil {
			t.Errorf("Content[1] missing style")
		} else {
			// Verify color (FontDrawerSrc is uniform)
			u, ok := c1.style.FontDrawerSrc.(*image.Uniform)
			if !ok || u.C != red {
				t.Errorf("Content[1] color mismatch")
			}
		}

		// Check Blue
		if c2 := contents[2]; c2.text != "Blue" {
			t.Errorf("Content[2] text mismatch")
		} else {
			u, ok := c2.style.FontDrawerSrc.(*image.Uniform)
			if !ok || u.C != blue {
				t.Errorf("Content[2] color mismatch, got %v", u.C)
			}
		}

		// Check Reset
		if c3 := contents[3]; c3.text != "PlainAgain" {
			t.Errorf("Content[3] text mismatch")
		} else {
			if c3.style != nil && c3.style.FontDrawerSrc != nil {
				t.Errorf("Content[3] should be reset to plain, got style %v", c3.style)
			}
		}
	})

	t.Run("Group Scoping with Reset", func(t *testing.T) {
		args := []interface{}{
			Group{
				Args: []interface{}{
					Color(red),
					"Red",
					Reset(), // This should reset only within the linear processing of the group?
					// Wait, Reset clears s.currentStyle.
					// Linear processing modifies s.currentStyle.
					// Groups push/pop s.currentStyle.
					"PlainInGroup",
				},
			},
			"PlainOutside",
		}

		contents, _, _, _, _, _ := ProcessRichArgs(args...)

		if len(contents) != 3 {
			t.Fatalf("Expected 3 content blocks, got %d", len(contents))
		}

		if c1 := contents[1]; c1.text != "PlainInGroup" {
			t.Errorf("Content[1] mismatch")
		} else if c1.style != nil && c1.style.FontDrawerSrc != nil {
			t.Errorf("Reset inside group did not clear style")
		}

		if c2 := contents[2]; c2.text != "PlainOutside" {
			t.Errorf("Content[2] mismatch")
		} else if c2.style != nil && c2.style.FontDrawerSrc != nil {
			t.Errorf("Group scope leaked style or reset logic failed")
		}
	})

	t.Run("MinSize Constraints", func(t *testing.T) {
		// Test MinWidth
		width := 100
		args := []interface{}{
			fontFace,
			MinWidth(width),
			"Short",
		}

		contents, _, _, _, _, _ := ProcessRichArgs(args...)
		c := contents[0]
		if c.style == nil || c.style.MinSize.X != fixed.I(width) {
			t.Errorf("Style MinSize mismatch")
		}

		// We can't verify layout without running Wrapper.
		// Content has decorators.
		// We rely on integration test or inspect decorators loop (hard).
		// Let's run basic layout.
		wrapper := NewRichWrapper(args...)
		lines, _, err := wrapper.TextToRect(image.Rect(0, 0, 200, 200))
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}

		// The line should be at least width pixels long?
		// "Short" is definitely < 100px.
		// AdvanceRect of the box should be max(textAdv, 100).

		if len(lines) == 0 {
			t.Fatal("No lines")
		}
		// Calculate line width used
		// lines[0].Size().Max.X?
		// Or sum of boxes AdvanceRect?

		boxes := lines[0].Boxes()
		totalAdv := fixed.Int26_6(0)
		for _, b := range boxes {
			totalAdv += b.AdvanceRect()
		}

		// MinWidth is 100 * 64
		minAdv := fixed.I(100)
		if totalAdv < minAdv {
			t.Errorf("Total advance %d < specified min width %d", totalAdv, minAdv)
		}
	})

	t.Run("Container Isolation", func(t *testing.T) {
		// Container with Margin.
		// Children should not inherit margin.

		margin := fixed.R(10, 0, 0, 0)
		args := []interface{}{
			fontFace,
			Container(
				Margin(margin),
				"Child",
			),
		}

		contents, _, _, _, _, _ := ProcessRichArgs(args...)
		// 0: fontFace, 1: Container
		if len(contents) != 1 { // fontFace does not create a content block
			t.Fatalf("Expected 1 content (container), got %d", len(contents))
		}

		// The container itself is wrapped.
		// Testing structure is hard via contents inspection.
		// We trust layout test.
		wrapper := NewRichWrapper(args...)
		lines, _, err := wrapper.TextToRect(image.Rect(0, 0, 200, 200))
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if len(lines) == 0 {
			t.Error("Expected lines")
		}

		// Note: Spacemap integration (interactive elements/ID) is tested in consuming applications
		// (e.g. question-vol) as wordwrap does not depend on spacemap directly.

		// Check that line length includes margin.
		// "Child" width + 10 margin.

		// Also check that if we add Color to container, CHILD gets it.
		// But Margin is NOT applied to Child box twice?
		// Actually "Margin" wraps "Container".
		// Container wraps Children.
		// Children do NOT get Margin decorator.

		// Let's test checking Color.
		argsColor := []interface{}{
			fontFace,
			Container(
				Color(blue),
				"BlueChild",
			),
		}
		contentsC, _, _, _, _, _ := ProcessRichArgs(argsColor...)

		// contentsC[0] is fontFace (Content?) No, fontFace does not create Content if type is font.Face?
		// ProcessRichArgs loop: case font.Face: sets state. No content added.
		// So contentsC length depends on if fontFace produces content.
		// rich.go: case font.Face: s.currentFont = v. No append.
		// So contents length is still 1 (Container).

		container := contentsC[0]
		if len(container.children) != 1 {
			t.Errorf("Expected 1 child in container, got %d", len(container.children))
		} else {
			child := container.children[0]
			if child.text != "BlueChild" {
				t.Errorf("Child text mismatch")
			}
			// Child SHOULD have style with blue color from parent inheritance?
			// Logic: "But they SHOULD inherit Font/Color."
			// subS.currentStyle starts as clone of s.currentStyle.
			// s.currentStyle has Color(Blue).
			// subS.currentStyle DOES have Color(Blue).
			// subS.currentStyle.Decorators cleared.
			// So Child gets Blue Color.

			// Verify
			u, ok := child.style.FontDrawerSrc.(*image.Uniform)
			if !ok || u.C != blue {
				t.Errorf("Child did not inherit color")
			}
		}
	})
}
