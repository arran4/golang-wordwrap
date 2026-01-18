package wordwrap_test

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
)

func Example_richTextComprehensive() {
	// 1. Setup Resources
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Fatal(err)
	}
	fontRegular := util.GetFontFace(24, 96, gr)
	fontLarge := util.GetFontFace(40, 96, gr)

	// Create a simple red box image for inline image demo
	redBox := image.NewRGBA(image.Rect(0, 0, 30, 30))
	draw.Draw(redBox, redBox.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	// Create a checker pattern for background demo
	pattern := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(pattern, pattern.Bounds(), &image.Uniform{color.RGBA{240, 240, 240, 255}}, image.Point{}, draw.Src)
	draw.Draw(pattern, image.Rect(0, 0, 5, 5), &image.Uniform{color.RGBA{200, 200, 200, 255}}, image.Point{}, draw.Src)
	draw.Draw(pattern, image.Rect(5, 5, 10, 10), &image.Uniform{color.RGBA{200, 200, 200, 255}}, image.Point{}, draw.Src)

	// 2. Build Rich Content
	// We mix strings with option structs. The order matters (state machine).
	args := []interface{}{
		fontRegular,
		"Standard text. ",

		// Text Color
		wordwrap.TextColor(color.RGBA{0, 0, 255, 255}),
		"Blue Text. ",
		wordwrap.TextColor(color.Black), // Reset

		// Background Color (Highlight)
		wordwrap.BgColor(color.RGBA{255, 255, 0, 255}),
		"Yellow Background. ",
		wordwrap.BgColor(color.Transparent), // Reset (or use Group)

		// Scoped Styles using Group
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.TextColor(color.RGBA{0, 100, 0, 255}),
				"Scoped Green Text. ",
				wordwrap.BgColor(color.RGBA{220, 255, 220, 255}),
				"Green on Light Green. ",
			},
		},
		"Back to Normal. ",
		"\n\n",

		// Font Size Changes
		fontLarge, "Large Text. ",
		fontRegular, "Normal Text. ",
		"\n\n",

		// Effects: Underline, Strikethrough
		"Text with ",
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.Underline(color.RGBA{255, 0, 0, 255}),
				"Red Underline",
			},
		},
		" and ",
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.Strikethrough(color.Black),
				"Strikethrough",
			},
		},
		".\n\n",

		// Inline Images and Alignment
		"Image aligned baseline: ",
		wordwrap.ImageContent{Image: redBox},
		" Text after.",
		"\n",
		"Image aligned Top: ",
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.Alignment(wordwrap.AlignTop),
				wordwrap.ImageContent{Image: redBox},
			},
		},
		" (Text Top).",
		"\n",
		"Image aligned Bottom: ",
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.Alignment(wordwrap.AlignBottom),
				wordwrap.ImageContent{Image: redBox},
			},
		},
		" (Text Bottom).",
		"\n\n",

		// Background Image Pattern
		wordwrap.Group{
			Args: []interface{}{
				wordwrap.BgImage(pattern),
				"Text on Pattern Background. ",
				fontLarge, "Even Large Text on Pattern.",
			},
		},
	}

	wrapper := wordwrap.NewRichWrapper(args...)

	// 3. Layout with standard page constraints
	result, err := wrapper.TextToSpecs(
		wordwrap.Width(wordwrap.Fixed(400)), // Narrow width to force wrapping
		wordwrap.Padding(20, color.Black),
		wordwrap.PageBackground(color.White),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Render
	img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))
	if result.PageBackground != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{result.PageBackground}, image.Point{}, draw.Src)
	}
	if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
		log.Fatal(err)
	}

	saveDocImage("richtext_comprehensive.png", img)

	// Output:
	// Generated doc/richtext_comprehensive.png
}
