package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
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
	gr, err := OpenFont(*fontname)
	if err != nil {
		log.Panicf("Error opening font %s: %s", *fontname, err)
	}
	grf := GetFontFace(*fontsize, *dpi, gr)
	grfd := &font.Drawer{
		Dst:  i,
		Src:  image.NewUniform(colornames.Blue),
		Face: grf,
	}
	text, err := GetText(*textsource)
	if err != nil {
		log.Panicf("Text fetch errror: %s", err)
	}
	ttb, _ := grfd.BoundString(text)
	grfd.Dot = grfd.Dot.Sub(ttb.Min)
	grfd.DrawString(text)

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

func GetFontFace(fontsize float64, dpi float64, gr *truetype.Font) font.Face {
	return truetype.NewFace(gr, &truetype.Options{
		Size: fontsize,
		DPI:  dpi,
	})
}

func OpenFont(name string) (*truetype.Font, error) {
	b, err := FontByName(name)
	if err != nil {
		return nil, fmt.Errorf("font open error: %w", err)
	}
	gr, err := truetype.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("font load error: %w", err)
	}
	return gr, nil
}

func FontByName(name string) ([]byte, error) {
	switch name {
	case "goregular":
		return goregular.TTF, nil
	}
	return nil, errors.New("font not found")
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
