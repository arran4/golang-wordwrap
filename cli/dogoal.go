package cli

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"path/filepath"

	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
)

//go:embed "goalimage.png"
var inputImage []byte

// DoGoal is a subcommand `wordwrap dogoal`
func DoGoal() error {
	log.Printf("Starting")
	ii, err := png.Decode(bytes.NewReader(inputImage))
	if err != nil {
		return fmt.Errorf("error opening image: %w", err)
	}

	i, ok := ii.(wordwrap.Image)
	if !ok {
		// If assertion fails, convert to RGBA
		b := ii.Bounds()
		dst := image.NewRGBA(b)
		draw.Draw(dst, b, ii, b.Min, draw.Src)
		i = dst
	}

	gr, err := util.OpenFont("goregular")
	if err != nil {
		return fmt.Errorf("error opening font %s: %w", "goregular", err)
	}
	grf := util.GetFontFace(16, 75, gr)
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus placerat fermentum quam aliquam lobortis."
	target := image.Rect(350, 44, 592, 209)
	sw := wordwrap.NewSimpleWrapper(text, grf)
	lines, _, err := sw.TextToRect(target)
	if err != nil {
		return fmt.Errorf("text wrap error: %w", err)
	}
	if err := sw.RenderLines(i, lines, target.Min); err != nil {
		return fmt.Errorf("text draw error: %w", err)
	}
	outfn := filepath.Join("images", "goalimage.png")
	if err := SaveFile(i, outfn); err != nil {
		return fmt.Errorf("error with saving file: %w", err)
	}
	log.Printf("Done as %s", outfn)
	return nil
}
