package wordwrap_test

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"

	"github.com/arran4/go-pattern"
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

	// Create a pebbles pattern for background demo using go-pattern
	// Using Scatter for true overlapping geometry.
	pat := pattern.NewScatter(
		pattern.SetBounds(image.Rect(-100000, -100000, 100000, 100000)),
		pattern.SetScatterFrequency(0.04), // Size control
		pattern.SetScatterDensity(1.0),    // Packed tight
		pattern.SetScatterMaxOverlap(1),
		pattern.SetSpaceColor(color.Transparent), // Transparent background!
		pattern.SetScatterGenerator(func(u, v float64, hash uint64) (color.Color, float64) {
			// Randomize size slightly
			rSize := float64(hash&0xFF) / 255.0
			radius := 12.0 + rSize*6.0 // 12 to 18 pixels radius

			// Perturb the shape using simple noise (simulated by sin/cos of hash+angle)
			// to make it "chipped" or irregular.
			angle := math.Atan2(v, u)
			dist := math.Sqrt(u*u + v*v)

			// Simple radial noise
			noise := math.Sin(angle*5+float64(hash%10)) * 0.1
			noise += math.Cos(angle*13+float64(hash%7)) * 0.05

			effectiveRadius := radius * (1.0 + noise)

			if dist > effectiveRadius {
				return color.Transparent, 0
			}

			// Stone Color: Lighter variations for readability
			grey := 220 + int(hash%35)
			col := color.RGBA{uint8(grey), uint8(grey), uint8(grey), 255}

			// Shading (diffuse)
			// Normal estimation for a flattened spheroid
			nx := u / effectiveRadius
			ny := v / effectiveRadius
			nz := math.Sqrt(math.Max(0, 1.0-nx*nx-ny*ny))

			// Light dir
			lx, ly, lz := -0.5, -0.5, 0.7
			lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
			lx, ly, lz = lx/lLen, ly/lLen, lz/lLen

			diffuse := math.Max(0, nx*lx+ny*ly+nz*lz)

			// Apply shading
			r := float64(col.R) * (0.1 + 0.9*diffuse)
			g := float64(col.G) * (0.1 + 0.9*diffuse)
			b := float64(col.B) * (0.1 + 0.9*diffuse)

			// Soft edge anti-aliasing
			alpha := 1.0
			edgeDist := effectiveRadius - dist
			if edgeDist < 1.0 {
				alpha = edgeDist
			}

			// Use hash for random Z-ordering
			z := float64(hash) / 18446744073709551615.0

			return color.RGBA{
				R: uint8(math.Min(255, r)),
				G: uint8(math.Min(255, g)),
				B: uint8(math.Min(255, b)),
				A: uint8(alpha * 255),
			}, z
		}),
	)

	// Create a paper-like pattern for the page background
	paper := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(paper, paper.Bounds(), &image.Uniform{color.RGBA{250, 250, 245, 255}}, image.Point{}, draw.Src)
	// Add some noise/lines to make it obvious
	for y := 0; y < 100; y += 20 {
		draw.Draw(paper, image.Rect(0, y, 100, y+1), &image.Uniform{color.RGBA{220, 220, 210, 255}}, image.Point{}, draw.Over)
	}

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
		// Demonstrate Resetting Background to Transparent (should show paper pattern)
		wordwrap.BgColor(color.Transparent),
		"Back to Normal (should see paper). ",
		// Optional: We can adds a helper for this later if needed
		// wordwrap.NoBackground(),

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
				wordwrap.BgImage(pat, wordwrap.BgPositioningPassThrough),
				"Text on Pattern Background. ",
				fontLarge, "Even Large Text on Pattern.",
			},
		},
	}

	wrapper := wordwrap.NewRichWrapper(args...)

	// 3. Layout with standard page constraints
	result, err := wrapper.TextToSpecs(
		wordwrap.Width(wordwrap.Fixed(800)), // Wider width as requested
		wordwrap.Padding(20, color.Black),
		// We don't set PageBackground here to avoid it overriding our manual draw if we relied on internal logic,
		// but since we draw manually below, it's fine.
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Render
	img := image.NewRGBA(image.Rect(0, 0, result.PageSize.X, result.PageSize.Y))

	// Draw Paper Pattern as background (Tiled)
	for y := 0; y < img.Bounds().Dy(); y += paper.Bounds().Dy() {
		for x := 0; x < img.Bounds().Dx(); x += paper.Bounds().Dx() {
			r := image.Rect(x, y, x+paper.Bounds().Dx(), y+paper.Bounds().Dy())
			draw.Draw(img, r, paper, image.Point{}, draw.Src)
		}
	}

	if err := wrapper.RenderLines(img, result.Lines, result.ContentStart); err != nil {
		log.Fatal(err)
	}

	saveDocImage("richtext_comprehensive.png", img)

	// Output:
	// Generated doc/richtext_comprehensive.png
}
