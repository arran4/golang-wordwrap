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
	width      = flag.Int("width", 400, "Doc width")
	height     = flag.Int("height", 600, "Doc height")
	dpi        = flag.Float64("dpi", 180, "Doc dpi")
	fontname   = flag.String("font", "goregular", "Text font")
	fontsize   = flag.Float64("size", 16, "font size")
	textsource = flag.String("text", "./testdata/sample1.txt", "File in, or - for std input")
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
		log.Panicf("Text fetch errror: %s", err)
	}
	n := 0
	rt := []rune(text)
	p := i.Rect.Min
	for {
		l, ni, err := wordwrap.SimpleFolder(wordwrap.SimpleBoxer, grf, rt[n:], i.Rect)
		if err != nil {
			log.Panicf("Error with boxing text: %s", err)
		}
		if l == nil {
			break
		}
		s := l.Size()
		rgba := i.SubImage(s.Add(p)).(*image.RGBA)
		if err := l.DrawLine(rgba); err != nil {
			log.Panicf("Error with drawing text: %s", err)
		}
		p.Y += s.Dy()
		n += ni

	}
	outfn := "out.png"
	if err := SaveFile(i, outfn); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Done as %s", outfn)
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
		return "", nil
	}
	return string(b), nil
}

func SaveFile(i *image.RGBA, fn string) error {
	os.MkdirAll("images", 0755)
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
