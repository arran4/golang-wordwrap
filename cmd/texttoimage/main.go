package main

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

func main() {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Fatal(err)
	}
	fontRegular := util.GetFontFace(24, 96, gr)
	fontLarge := util.GetFontFace(48, 96, gr)

	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	longLorem := lorem + " " + lorem + " " + lorem

	// Define Test Cases using the new TextToImage API
	tests := []struct {
		Name string
		Args []interface{}
	}{
		{
			Name: "01_Unbounded_Default",
			Args: []interface{}{
				fontRegular,
				"Default: Unbounded width and height. Transparent Background.",
				longLorem,
			},
		},
		{
			Name: "02_Fixed_Width_A4",
			Args: []interface{}{
				fontRegular,
				"Fixed A4 Width (794px). Height Unbounded. White Background.",
				"\n\n",
				longLorem, longLorem,
				wordwrap.Width(wordwrap.Fixed(794)),
				wordwrap.PageBackground(color.White),
			},
		},
		{
			Name: "03_Min_A4_Unbounded",
			Args: []interface{}{
				fontRegular,
				"Min(A4, Unbounded). White Background.",
				"\n\n",
				"Short line.",
				"\n\n",
				longLorem, longLorem,
				wordwrap.Width(wordwrap.Min(wordwrap.A4Width(96), wordwrap.Unbounded())),
				wordwrap.PageBackground(color.White),
			},
		},
		{
			Name: "04_Small_Min_Constraint",
			Args: []interface{}{
				fontRegular,
				"Min(Fixed(300), Unbounded). Capped at 300px. White Background.",
				"\n\n",
				longLorem,
				wordwrap.Width(wordwrap.Min(wordwrap.Fixed(300), wordwrap.Unbounded())),
				wordwrap.PageBackground(color.White),
			},
		},
		{
			Name: "05_Max_Constraint",
			Args: []interface{}{
				fontRegular,
				"Max(Fixed(300), Unbounded). At least 300px. White Background.",
				"\n\n",
				"Tiny.",
				wordwrap.Width(wordwrap.Max(wordwrap.Fixed(300), wordwrap.Unbounded())),
				wordwrap.PageBackground(color.White),
			},
		},
		{
			Name: "06_Complex_Rich_Text",
			Args: []interface{}{
				fontRegular,
				wordwrap.TextColor(color.RGBA{255, 0, 0, 255}), "Red Header\n",
				fontLarge, "Large Content\n",
				fontRegular,
				wordwrap.Width(wordwrap.Min(wordwrap.Fixed(400), wordwrap.Unbounded())),
				"Wrapped at 400px. Transparent Background.",
				longLorem,
			},
		},
		{
			Name: "07_Margin_Padding",
			Args: []interface{}{
				fontRegular,
				"Document with 20px padding/margin. White Background.",
				"\n\n",
				longLorem,
				wordwrap.Width(wordwrap.Fixed(400)), // Content width will be 400 - 40 = 360
				wordwrap.Padding(20, color.Black),   // SpecOption
				wordwrap.PageBackground(color.White),
			},
		},
		{
			Name: "08_HighDPI_Margin",
			Args: []interface{}{
				util.GetFontFace(24, 300, gr),
				"High DPI Document (300 DPI) with Padding. White Background.",
				"\n\n",
				longLorem, longLorem, longLorem,
				wordwrap.Width(wordwrap.Min(wordwrap.A4Width(300), wordwrap.Unbounded())),
				wordwrap.Padding(200, color.Black),
				wordwrap.PageBackground(color.White),
			},
		},
	}

	for _, tc := range tests {
		fmt.Printf("Generating %s...\n", tc.Name)

		// Separate Wrapper Args from Spec Args
		var wrapperArgs []interface{}
		var specOpts []wordwrap.SpecOption

		for _, arg := range tc.Args {
			if opt, ok := arg.(wordwrap.SpecOption); ok {
				specOpts = append(specOpts, opt)
			} else {
				wrapperArgs = append(wrapperArgs, arg)
			}
		}

		wrapper := wordwrap.NewRichWrapper(wrapperArgs...)
		result, err := wrapper.TextToSpecs(specOpts...)
		if err != nil {
			log.Printf("Error layouts %s: %v", tc.Name, err)
			continue
		}

		img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))

		// Handle Page Background
		if result.PageBackground != nil {
			draw.Draw(img, img.Bounds(), &image.Uniform{result.PageBackground}, image.Point{}, draw.Src)
		}

		if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
			log.Printf("Error rendering %s: %v", tc.Name, err)
			continue
		}

		f, err := os.Create(fmt.Sprintf("%s.png", tc.Name))
		if err != nil {
			log.Printf("Error creating file: %v", err)
			continue
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}()
		if err := png.Encode(f, img); err != nil {
			log.Printf("Error encoding png: %v", err)
		}
		fmt.Printf("Saved %s.png (%dx%d)\n", tc.Name, img.Bounds().Dx(), img.Bounds().Dy())
	}
}
