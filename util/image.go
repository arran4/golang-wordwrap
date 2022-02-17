package util

import (
	"image"
	"image/color"
	"image/draw"
)

func DrawBox(i draw.Image, s image.Rectangle) {
	for x := s.Min.X; x < s.Max.X; x++ {
		i.Set(x, s.Min.Y, color.Black)
		i.Set(x, s.Max.Y-1, color.Black)
	}
	for y := s.Min.Y; y < s.Max.Y; y++ {
		i.Set(s.Min.X, y, color.Black)
		i.Set(s.Max.X-1, y, color.Black)
	}
}
