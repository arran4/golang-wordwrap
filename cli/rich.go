package cli

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goitalic"
)

const A4Width = 794 // 210mm at 96 DPI

// RichWrapToImage is a subcommand `wordwrap rich`
//
// Flags:
//
//	width:  --width  (default: 0) Doc width (0 = auto/A4)
//	height: --height (default: 0) Doc height (0 = auto/unbounded)
//	out:    --out    (default: "rich_output.png") Output filename
func RichWrapToImage(width int, height int, out string) error {
	// 2. Load Fonts (Using goregular via util)
	gr, err := util.OpenFont("goregular")
	if err != nil {
		return fmt.Errorf("failed to open font: %w", err)
	}

	// Load Italic Font
	gi, err := truetype.Parse(goitalic.TTF)
	if err != nil {
		return fmt.Errorf("failed to parse italic font: %w", err)
	}

	// Define variants
	fontRegular := util.GetFontFace(24, 96, gr)
	fontLarge := util.GetFontFace(48, 96, gr)
	fontSmall := util.GetFontFace(16, 96, gr)
	fontItalic := truetype.NewFace(gi, &truetype.Options{
		Size:    24,
		DPI:     96,
		Hinting: font.HintingNone,
	})

	// 3. Create Assets
	// Textures (Local implementations to avoid external dependency issues if any)
	checkered := CheckerPattern(color.RGBA{220, 220, 220, 255}, color.White, 10)
	striped := StripePattern(color.RGBA{200, 200, 255, 255}, color.White, 10)

	// A red square image
	redSquare := image.NewRGBA(image.Rect(0, 0, 40, 40))
	draw.Draw(redSquare, redSquare.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	// A blue circle-ish image
	blueRect := image.NewRGBA(image.Rect(0, 0, 30, 30))
	draw.Draw(blueRect, blueRect.Bounds(), &image.Uniform{color.RGBA{0, 0, 255, 255}}, image.Point{}, draw.Src)

	// Function to create wrapper (factory)
	createWrapper := func() *wordwrap.SimpleWrapper {
		return wordwrap.NewRichWrapper(
			fontRegular,
			"Welcome to the ",
			wordwrap.Group{Args: []interface{}{
				wordwrap.TextColor(color.RGBA{200, 0, 0, 255}), "Rich Text ",
				wordwrap.TextColor(color.Black), "demonstration with groups!\n\n",
			}},

			"We can have ", fontLarge, "Large Text", fontRegular, " and ", fontSmall, "tiny text", fontRegular, " inline.\n\n",

			"Highlights and Effects:\n",
			wordwrap.Group{Args: []interface{}{
				"Standard Highlight: ", wordwrap.Highlight(color.RGBA{255, 255, 0, 255}), "Yellow Highlight", fontRegular, "\n",
				"Strikethrough: ", wordwrap.Strikethrough(color.Black), "deleted text", fontRegular, "\n",
				"Underline: ", wordwrap.Underline(color.Black), "underlined text", fontRegular, "\n",
			}},

			"Line positioning (Alignment):\n",
			wordwrap.Group{Args: []interface{}{
				"Baseline: ", wordwrap.Alignment(wordwrap.AlignBaseline), "Base ",
				"Top: ", wordwrap.Alignment(wordwrap.AlignTop), fontSmall, "Top ", fontRegular,
				"Middle: ", wordwrap.Alignment(wordwrap.AlignMiddle), fontLarge, "Mid ", fontRegular,
				"Bottom: ", wordwrap.Alignment(wordwrap.AlignBottom), fontSmall, "Bot ", fontRegular,
				"\n\n",
			}},

			"Background Textures and Images:\n",
			wordwrap.BgImage(checkered), "Checkered Background Text ",
			wordwrap.BgImage(striped), "Striped Background Text\n\n",

			"Italics (Real font loaded):\n",
			"This is regular. ",
			wordwrap.Group{Args: []interface{}{
				fontItalic, // Use actual italic font
				"This is true italic style text. ",
			}},
			fontRegular, // Switch back
			"Back to regular.\n\n",

			"Images with Alignment:\n",
			"Inline: ",
			wordwrap.Alignment(wordwrap.AlignMiddle), wordwrap.ImageContent{Image: redSquare},
			wordwrap.Alignment(wordwrap.AlignBaseline), " (Middle aligned image) and ",
			wordwrap.Alignment(wordwrap.AlignTop), wordwrap.ImageContent{Image: blueRect},
			wordwrap.Alignment(wordwrap.AlignBaseline), " (Top aligned image).\n",
		)
	}

	// Calculate Dimensions
	targetWidth := width
	targetHeight := height

	if targetWidth <= 0 {
		// Pass 1: Calculate natural width
		w := createWrapper()
		lines, _, err := w.TextToRect(image.Rect(0, 0, 100000, 100000))
		if err != nil {
			return fmt.Errorf("auto-sizing Pass 1 error: %w", err)
		}
		maxLineWidth := 0
		for _, l := range lines {
			s := l.Size()
			if s.Dx() > maxLineWidth {
				maxLineWidth = s.Dx()
			}
		}
		if maxLineWidth > A4Width {
			targetWidth = A4Width
		} else {
			targetWidth = maxLineWidth
		}
		// Ensure non-zero
		if targetWidth == 0 {
			targetWidth = 1
		}
	}

	// Pass 2: Layout with determining height (if auto)
	finalWrapper := createWrapper()
	layoutHeight := targetHeight
	if layoutHeight <= 0 {
		layoutHeight = 100000 // Huge height
	}

	lines, p, err := finalWrapper.TextToRect(image.Rect(0, 0, targetWidth, layoutHeight))
	if err != nil {
		return fmt.Errorf("layout error: %w", err)
	}

	if targetHeight <= 0 {
		targetHeight = p.Y
	}

	// Create Image
	img := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	if err := finalWrapper.RenderLines(img, lines, img.Bounds().Min); err != nil {
		return fmt.Errorf("render error: %w", err)
	}

	// 6. Save
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	fmt.Printf("Rich text image generated at: %s (%dx%d)\n", out, targetWidth, targetHeight)
	return nil
}

func CheckerPattern(c1, c2 color.Color, size int) image.Image {
	i := image.NewRGBA(image.Rect(0, 0, size*2, size*2))
	draw.Draw(i, i.Bounds(), &image.Uniform{c1}, image.Point{}, draw.Src)
	draw.Draw(i, image.Rect(0, 0, size, size), &image.Uniform{c2}, image.Point{}, draw.Src)
	draw.Draw(i, image.Rect(size, size, size*2, size*2), &image.Uniform{c2}, image.Point{}, draw.Src)
	return TiledImage{I: i}
}

func StripePattern(c1, c2 color.Color, size int) image.Image {
	i := image.NewRGBA(image.Rect(0, 0, size, size*2))
	draw.Draw(i, i.Bounds(), &image.Uniform{c1}, image.Point{}, draw.Src)
	draw.Draw(i, image.Rect(0, size, size, size*2), &image.Uniform{c2}, image.Point{}, draw.Src)
	return TiledImage{I: i}
}

// TiledImage tiling image
type TiledImage struct {
	I image.Image
}

func (t TiledImage) ColorModel() color.Model {
	return t.I.ColorModel()
}

func (t TiledImage) Bounds() image.Rectangle {
	return image.Rect(-1e9, -1e9, 1e9, 1e9)
}

func (t TiledImage) At(x, y int) color.Color {
	b := t.I.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return color.Transparent
	}
	xMod := x % w
	if xMod < 0 {
		xMod += w
	}
	yMod := y % h
	if yMod < 0 {
		yMod += h
	}
	return t.I.At(b.Min.X+xMod, b.Min.Y+yMod)
}
