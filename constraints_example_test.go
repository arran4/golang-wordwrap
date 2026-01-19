package wordwrap_test

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
)

// Helper to save image for documentation
func saveDocImage(name string, img image.Image) {
	f, err := os.Create("doc/" + name)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated doc/%s\n", name)
}

func ExampleSimpleWrapper_TextToSpecs_simple() {
	// Open Font
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Fatal(err)
	}
	font := util.GetFontFace(24, 96, gr)

	// Standard Wrapper Args
	args := []interface{}{
		font,
		"Simple Text wrapping example.",
	}

	// Create Wrapper
	wrapper := wordwrap.NewRichWrapper(args...)

	// Layout with Constraint: Fixed Width 200px
	result, err := wrapper.TextToSpecs(wordwrap.Width(wordwrap.Fixed(200)))
	if err != nil {
		log.Fatal(err)
	}

	// Render
	img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))
	if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
		log.Fatal(err)
	}

	saveDocImage("simple_example.png", img)

	// Output:
	// Generated doc/simple_example.png
}

func ExampleSimpleWrapper_TextToSpecs_a4() {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Fatal(err)
	}
	font := util.GetFontFace(24, 96, gr)

	text := "This is an example of an A4 document layout. It uses a fixed A4 width and standard padding."

	wrapper := wordwrap.NewRichWrapper(
		font,
		text,
	)

	// Layout Specs: A4 Width (96 DPI), 20px Padding, White Background
	result, err := wrapper.TextToSpecs(
		wordwrap.Width(wordwrap.A4Width(96)),
		wordwrap.Padding(20, color.Black),
		wordwrap.PageBackground(color.White),
	)
	if err != nil {
		log.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))

	// Draw Background
	if result.PageBackground != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{result.PageBackground}, image.Point{}, draw.Src)
	}

	if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
		log.Fatal(err)
	}

	saveDocImage("a4_example.png", img)

	// Output:
	// Generated doc/a4_example.png
}

func ExampleSimpleWrapper_TextToSpecs_flexible() {
	// Example of "Min(300, Unbounded)"
	// This creates a box that is at least 300px wide, but grows if content is wider (unlikely for text wrapping,
	// but useful if 'Unbounded' is default and you want a minimum).
	// Actually, "Min(A4, Unbounded)" means "Min(A4, Natural)".
	// If Natural < A4, use Natural? NO. Min(A, B) uses smaller.
	// Natural is usually small for short text. A4 is large.
	// So Min(A4, Natural) = Natural.
	// This implements "Auto width but max A4".

	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Fatal(err)
	}
	font := util.GetFontFace(24, 96, gr)

	// Long text to demonstrate wrapping
	wrapper := wordwrap.NewRichWrapper(
		font,
		"This text will wrap at A4 width because of the constraint logic.",
	)

	result, err := wrapper.TextToSpecs(
		wordwrap.Width(wordwrap.Min(wordwrap.A4Width(96), wordwrap.Unbounded())),
		wordwrap.PageBackground(color.White),
	)
	if err != nil {
		log.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))
	if result.PageBackground != nil {
		draw.Draw(img, img.Bounds(), &image.Uniform{result.PageBackground}, image.Point{}, draw.Src)
	}
	if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
		log.Fatal(err)
	}

	saveDocImage("flexible_example.png", img)

	// Output:
	// Generated doc/flexible_example.png
}
