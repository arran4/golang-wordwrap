package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var (
	//go:embed "goalimage.png"
	inputImage []byte
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	flag.Parse()
	log.Printf("Starting")
	ii, err := png.Decode(bytes.NewReader(inputImage))
	if err != nil {
		log.Panicf("Error opening image: %s", err)
	}
	i := ii.(wordwrap.Image)
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Panicf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 75, gr)
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus placerat fermentum quam aliquam lobortis."
	target := image.Rect(350, 44, 592, 209)
	sw, lines, _, err := wordwrap.SimpleWrapTextToRect(target, nil, wordwrap.NewContent(text, wordwrap.WithFont(grf)))
	if err != nil {
		log.Panicf("Text wrap error: %s", err)
	}
	if err := sw.RenderLines(i, lines, target.Min); err != nil {
		log.Panicf("Text draw error: %s", err)
	}
	outfn := filepath.Join("images", "goalimage.png")
	if err := SaveFile(i, outfn); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Done as %s", outfn)
}

func SaveFile(i wordwrap.Image, fn string) error {
	_ = os.MkdirAll("images", 0755)
	fi, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Printf("File close error: %s", err)
		}
	}()
	if err := png.Encode(fi, i); err != nil {
		return fmt.Errorf("png encoding: %w", err)
	}
	return nil
}
