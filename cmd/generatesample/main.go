package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

var (
	//go:embed "chevron.png"
	chevronImageBytes []byte
)

func main() {
	flag.Parse()
	chevronImage, _ := png.Decode(bytes.NewReader(chevronImageBytes))
	grf, err := GetFontFace("goregular", 16, 180)
	if err != nil {
		log.Panicf("Error opening font %s: %s", "goregular", err)
	}
	fontDrawer := &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: grf,
	}
	SampleType1(400, 600, 16, 180, "goregular", "testdata/sample1.txt", "images/sample01.png")
	SampleType1(400, 600, 16, 180, "goregular", "testdata/sample2.txt", "images/sample02.png", wordwrap.BoxLine)
	SampleType1(400, 600, 16, 180, "goregular", "testdata/sample3.txt", "images/sample03.png")
	SampleType1(400, 600, 16, 180, "goregular", "testdata/sample4.txt", "images/sample04.png", wordwrap.BoxBox)
	SampleType1(400, 85, 16, 180, "goregular", "testdata/sample5.txt", "images/sample05.png", wordwrap.YOverflow(wordwrap.DescentOverflow))
	SampleType1(200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample06.png", wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage)))
	SampleType1(200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample07.png", wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricAboveTheLine), wordwrap.BoxBox))
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample1.txt", "images/sample08.png", wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricBelowTheLine), wordwrap.BoxBox))
	SampleType1(200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample09.png", wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricCenter(fontDrawer)), wordwrap.BoxBox))
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample1.txt", "images/sample10.png", wordwrap.HorizontalCenterLines)
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample1.txt", "images/sample11.png", wordwrap.RightLines)
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample7.txt", "images/sample12.png", wordwrap.HorizontalCenterBlock)
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample7.txt", "images/sample13.png", wordwrap.RightBlock)
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample6.txt", "images/sample14.png", wordwrap.VerticalCenterBlock)
	SampleType1(200, 120, 16, 75, "goregular", "testdata/sample6.txt", "images/sample15.png", wordwrap.BottomBlock)
}

func SampleType1(width, height int, fontsize, dpi float64, fontname, textsource, outfilename string, opts ...wordwrap.WrapperOption) {
	log.Printf("Working on %s", outfilename)
	i := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(i, i.Bounds(), image.NewUniform(image.White), i.Bounds().Min, draw.Over)
	grf, err := GetFontFace(fontname, fontsize, dpi)
	if err != nil {
		log.Panicf("Error opening font %s: %s", fontname, err)
	}
	text, err := GetText(textsource)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	if err := wordwrap.SimpleWrapTextToImage(text, i, grf, opts...); err != nil {
		log.Panicf("Text wrap and draw error: %s", err)
	}
	if err := SaveFile(i, outfilename); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Done as %s", outfilename)
}

func GetFontFace(fontname string, fontsize float64, dpi float64) (font.Face, error) {
	gr, err := util.OpenFont(fontname)
	if err != nil {
		return nil, err
	}
	grf := util.GetFontFace(fontsize, dpi, gr)
	return grf, err
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
