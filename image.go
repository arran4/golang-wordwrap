package wordwrap

import (
	"image"
	"image/draw"
)

type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}
