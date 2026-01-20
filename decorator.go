package wordwrap

import (
	"image"
	"image/draw"

	"golang.org/x/image/math/fixed"
)

// DecorationBox is a box that adds padding and margin around another box
type DecorationBox struct {
	Box
	Padding       fixed.Rectangle26_6
	Margin        fixed.Rectangle26_6
	Background    image.Image
	BgPositioning BackgroundPositioning
}

// NewDecorationBox constructor
func NewDecorationBox(b Box, padding, margin fixed.Rectangle26_6, bg image.Image, bgPos BackgroundPositioning) *DecorationBox {
	return &DecorationBox{
		Box:           b,
		Padding:       padding,
		Margin:        margin,
		Background:    bg,
		BgPositioning: bgPos,
	}
}

// DrawBox renders the box with decorations.
func (db *DecorationBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	b := i.Bounds()

	// Margin is outside the background.
	// Padding is inside the background.

	bgRect := b
	bgRect.Min.X += db.Margin.Min.X.Ceil()
	bgRect.Min.Y += db.Margin.Min.Y.Ceil()
	bgRect.Max.X -= db.Margin.Max.X.Ceil()
	bgRect.Max.Y -= db.Margin.Max.Y.Ceil()

	// Inner Box Rect (Content Box)
	innerRect := bgRect
	innerRect.Min.X += db.Padding.Min.X.Ceil()
	innerRect.Min.Y += db.Padding.Min.Y.Ceil()
	innerRect.Max.X -= db.Padding.Max.X.Ceil()
	innerRect.Max.Y -= db.Padding.Max.Y.Ceil()

	if db.Background != nil {
		srcPoint := image.Point{}
		switch db.BgPositioning {
		case BgPositioningPassThrough:
			srcPoint = bgRect.Min
		case BgPositioningZeroed:
			srcPoint = image.Point{}
		case BgPositioningSection5Zeroed:
			srcPoint = bgRect.Min.Sub(innerRect.Min)
		}
		draw.Draw(i, bgRect, db.Background, srcPoint, draw.Over)
	}

	if innerRect.Empty() {
		// No space for inner box.
		return
	}

	innerImg := i.SubImage(innerRect).(Image)

	yOffset := db.Padding.Min.Y + db.Margin.Min.Y
	db.Box.DrawBox(innerImg, y-yOffset, dc)
}

func (db *DecorationBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	w, h := db.Box.MinSize()
	return w + db.Padding.Max.X + db.Padding.Min.X + db.Margin.Max.X + db.Margin.Min.X,
		h + db.Padding.Max.Y + db.Padding.Min.Y + db.Margin.Max.Y + db.Margin.Min.Y
}

func (db *DecorationBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	w, h := db.Box.MaxSize()
	if w == 0 && h == 0 {
		return 0, 0
	}
	return w + db.Padding.Max.X + db.Padding.Min.X + db.Margin.Max.X + db.Margin.Min.X,
		h + db.Padding.Max.Y + db.Padding.Min.Y + db.Margin.Max.Y + db.Margin.Min.Y
}
