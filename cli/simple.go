package cli

import (
	"fmt"
	"image"
	"log"
	"strconv"

	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
)

// SimpleWrapToImage is a subcommand `wordwrap simple`
//
// Flags:
//
//	width:       --width       (default: 400)         Doc width
//	height:      --height      (default: 600)         Doc height
//	dpiStr:      --dpi         (default: "180.0")     Doc dpi
//	fontname:    --font        (default: "goregular") Text font
//	fontsizeStr: --size        (default: "16.0")      font size
//	textsource:  --text        (default: "-")         File in, or - for std input
//	outfilename: --out         (default: "out.png")   file to write to, in some cases this is ignored
//	boxline:     --boxline     (default: false)       Box the line
//	boxbox:      --boxbox      (default: false)       Box the box
//	yoverflow:   --yoverflow   (default: 0)           Y Overflow mode
func SimpleWrapToImage(width int, height int, dpiStr string, fontname string, fontsizeStr string, textsource string, outfilename string, boxline bool, boxbox bool, yoverflow int) error {
	dpi, err := strconv.ParseFloat(dpiStr, 64)
	if err != nil {
		return fmt.Errorf("invalid dpi: %w", err)
	}
	fontsize, err := strconv.ParseFloat(fontsizeStr, 64)
	if err != nil {
		return fmt.Errorf("invalid fontsize: %w", err)
	}

	i := image.NewRGBA(image.Rect(0, 0, width, height))
	gr, err := util.OpenFont(fontname)
	if err != nil {
		return fmt.Errorf("error opening font %s: %w", fontname, err)
	}
	grf := util.GetFontFace(fontsize, dpi, gr)
	text, err := GetText(textsource)
	if err != nil {
		return fmt.Errorf("text fetch error: %w", err)
	}
	var opts []wordwrap.WrapperOption
	if boxline {
		opts = append(opts, wordwrap.BoxLine)
	}
	if boxbox {
		opts = append(opts, wordwrap.BoxBox)
	}
	if yoverflow > 0 {
		opts = append(opts, wordwrap.YOverflow(wordwrap.OverflowMode(yoverflow)))
	}
	sw := wordwrap.NewSimpleWrapper(text, grf, opts...)
	lines, _, err := sw.TextToRect(i.Bounds())
	if err != nil {
		return fmt.Errorf("text wrap error: %w", err)
	}
	if err := sw.RenderLines(i, lines, i.Bounds().Min); err != nil {
		return fmt.Errorf("text draw error: %w", err)
	}
	if err := SaveFile(i, outfilename); err != nil {
		return fmt.Errorf("error with saving file: %w", err)
	}
	log.Printf("Done as %s", outfilename)
	return nil
}
