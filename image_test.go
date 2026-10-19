package wordwrap

import (
	"image"
	"image/color"
	"testing"
)

func TestTiledImage(t *testing.T) {
	t.Run("Zero width or height bounds", func(t *testing.T) {
		testCases := []struct {
			name string
			rect image.Rectangle
		}{
			{"0x0", image.Rect(0, 0, 0, 0)},
			{"0xN", image.Rect(0, 0, 0, 10)},
			{"Nx0", image.Rect(0, 0, 10, 0)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				emptyImg := image.NewRGBA(tc.rect)
				tiled := Tiled(emptyImg)

				// This should not panic
				c := tiled.At(0, 0)
				_, _, _, a := c.RGBA()
				if a != 0 {
					t.Errorf("Expected transparent color, got %v", c)
				}
			})
		}
	})

	t.Run("Tiling behavior", func(t *testing.T) {
		// Create a 10x10 image with origin at (5, 5)
		// Bounds are Min: (5,5), Max: (15,15)
		img := image.NewRGBA(image.Rect(5, 5, 15, 15))

		// Set distinct colors at known locations within the source image bounds
		// (5,5) is the top-left of the source image
		img.Set(5, 5, color.RGBA{255, 0, 0, 255}) // Red
		// (14,14) is the bottom-right of the source image
		img.Set(14, 14, color.RGBA{0, 255, 0, 255}) // Green

		tiled := Tiled(img)

		testCases := []struct {
			name     string
			query    image.Point
			expected color.Color
		}{
			{
				name:     "Within bounds top-left maps to top-left",
				query:    image.Point{5, 5},
				expected: color.RGBA{255, 0, 0, 255},
			},
			{
				name:     "Within bounds bottom-right maps to bottom-right",
				query:    image.Point{14, 14},
				expected: color.RGBA{0, 255, 0, 255},
			},
			{
				name: "Negative coordinates map to correct modulo",
				// x = -5. (-5 - 5) % 10 = -10 % 10 = 0. + 5 = 5.
				// y = -5. (-5 - 5) % 10 = -10 % 10 = 0. + 5 = 5.
				query:    image.Point{-5, -5},
				expected: color.RGBA{255, 0, 0, 255},
			},
			{
				name: "Origin query (0,0) maps to center of image",
				// x = 0. (0 - 5) % 10 = -5. -5 < 0 -> 5. 5 + 5 = 10.
				// y = 0. (0 - 5) % 10 = -5. -5 < 0 -> 5. 5 + 5 = 10.
				query:    image.Point{0, 0},
				expected: color.RGBA{0, 0, 0, 0}, // Transparent, since (10,10) was not set.
			},
			{
				name: "Query right before origin maps to bottom-right",
				// x = 4. (4 - 5) % 10 = -1. -1 < 0 -> 9. 9 + 5 = 14.
				// y = 4. (4 - 5) % 10 = -1. -1 < 0 -> 9. 9 + 5 = 14.
				query:    image.Point{4, 4},
				expected: color.RGBA{0, 255, 0, 255},
			},
			{
				name: "Query far outside bounds",
				// x = 25. (25 - 5) % 10 = 20 % 10 = 0. + 5 = 5.
				// y = 25. (25 - 5) % 10 = 20 % 10 = 0. + 5 = 5.
				query:    image.Point{25, 25},
				expected: color.RGBA{255, 0, 0, 255},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := tiled.At(tc.query.X, tc.query.Y)

				gr, gg, gb, ga := got.RGBA()
				er, eg, eb, ea := tc.expected.RGBA()

				if gr != er || gg != eg || gb != eb || ga != ea {
					t.Errorf("At(%d, %d) = %v; want %v", tc.query.X, tc.query.Y, got, tc.expected)
				}
			})
		}
	})
}
