package wordwrap

import (
	"image"
	"image/color"
	"testing"
)

func TestTiledImage(t *testing.T) {
	t.Run("Zero width or height", func(t *testing.T) {
		emptyImg := image.NewRGBA(image.Rect(0, 0, 0, 0))
		tiled := Tiled(emptyImg)

		// This should not panic
		c := tiled.At(0, 0)
		_, _, _, a := c.RGBA()
		if a != 0 {
			t.Errorf("Expected transparent color, got %v", c)
		}
	})

	t.Run("Negative coordinates", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		img.Set(9, 9, color.White)

		tiled := Tiled(img)

		// -1, -1 should map to 9, 9
		c := tiled.At(-1, -1)
		r, g, b, a := c.RGBA()
		if r != 65535 || g != 65535 || b != 65535 || a != 65535 {
			t.Errorf("Expected white color at -1, -1, got %v", c)
		}
	})

	t.Run("Non-zero source origin", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(5, 5, 15, 15))
		img.Set(5, 5, color.White)

		tiled := Tiled(img)

		// 0, 0 should map to 5, 5 (since 0 % 10 = 0 + 5) - wait,
		// (0 - 5) % 10 = -5 -> +10 = 5.
		// b.Min.X + x = 5 + 5 = 10. Wait.

		// Let's test the At function logic:
		// x = 0
		// w = 10
		// x = (0 - 5) % 10 = -5
		// if x < 0: x += 10 => x = 5
		// At: src.At(5 + 5, 5 + 5) = src.At(10, 10)

		c := tiled.At(10, 10)
		// x = 10
		// x = (10 - 5) % 10 = 5 % 10 = 5
		// src.At(5 + 5) = src.At(10)

		// Wait, 5,5 is at x=0
		c = tiled.At(0, 0)
		r, g, b, a := c.RGBA()
		if r == 65535 { // It's white at 5, 5? But mapped to 10,10.
			t.Errorf("Expected non-white color at 0, 0, got %v", c)
		}

		// Map back to 5,5:
		// x = 5. (5-5)%10 = 0. src.At(5+0) = src.At(5).
		c = tiled.At(5, 5)
		r, g, b, a = c.RGBA()
		if r != 65535 || g != 65535 || b != 65535 || a != 65535 {
			t.Errorf("Expected white color at 5, 5, got %v", c)
		}

		// Mapped from -5, -5
		c = tiled.At(-5, -5)
		r, g, b, a = c.RGBA()
		if r != 65535 || g != 65535 || b != 65535 || a != 65535 {
			t.Errorf("Expected white color at -5, -5, got %v", c)
		}
	})
}
