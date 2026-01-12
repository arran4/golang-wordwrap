package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	width       = flag.Int("width", 400, "Doc width")
	height      = flag.Int("height", 600, "Doc height")
	dpi         = flag.Float64("dpi", 180, "Doc dpi")
	fontname    = flag.String("font", "goregular", "Text font")
	fontsize    = flag.Float64("size", 16, "font size")
	textsource  = flag.String("text", "-", "File in, or - for std input")
	outfilename = flag.String("out", "out.png", "file to write to, in some cases this is ignored")
	boxline     = flag.Bool("boxline", false, "Box the line")
	boxbox      = flag.Bool("boxbox", false, "Box the box")
	yoverflow   = flag.Int("yoverflow", 0, "Y Overflow mode")
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	flag.Parse()
	i := image.NewRGBA(image.Rect(0, 0, *width, *height))
	gr, err := util.OpenFont(*fontname)
	if err != nil {
		log.Panicf("Error opening font %s: %s", *fontname, err)
	}
	grf := util.GetFontFace(*fontsize, *dpi, gr)
	text, err := GetText(*textsource)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	var opts []wordwrap.WrapperOption
	if *boxline {
		opts = append(opts, wordwrap.BoxLine)
	}
	if *boxbox {
		opts = append(opts, wordwrap.BoxBox)
	}
	if *yoverflow > 0 {
		opts = append(opts, wordwrap.YOverflow(wordwrap.OverflowMode(*yoverflow)))
	}
	if err := wordwrap.SimpleWrapTextToImage(i, opts, wordwrap.NewContent(text, wordwrap.WithFont(grf))); err != nil {
		log.Panicf("Text wrap and draw error: %s", err)
	}
	if err := SaveFile(i, *outfilename); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Done as %s", *outfilename)
}

func GetText(fn string) (string, error) {
	if fn == "" {
		return "", errors.New("no input file specified")
	}
	if fn == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", fn, err)
	}
	return string(b), nil
}

func SaveFile(i *image.RGBA, fn string) error {
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
