package cli

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"

	"github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
)

//go:embed "chevron.png"
var chevronImageBytes []byte

// GenerateSample is a subcommand `wordwrap sample`
func GenerateSample() error {
	chevronImage, _ := png.Decode(bytes.NewReader(chevronImageBytes))
	grf, err := getFontFace("goregular", 16, 180)
	if err != nil {
		return fmt.Errorf("Error opening font %s: %w", "goregular", err)
	}
	fontDrawer := &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: grf,
	}

	type sample struct {
		w, h      int
		size, dpi float64
		font      string
		src, dst  string
		opts      []wordwrap.WrapperOption
	}

	samples := []sample{
		{400, 600, 16, 180, "goregular", "testdata/sample1.txt", "images/sample01.png", nil},
		{400, 600, 16, 180, "goregular", "testdata/sample2.txt", "images/sample02.png", []wordwrap.WrapperOption{wordwrap.BoxLine}},
		{400, 600, 16, 180, "goregular", "testdata/sample3.txt", "images/sample03.png", nil},
		{400, 600, 16, 180, "goregular", "testdata/sample4.txt", "images/sample04.png", []wordwrap.WrapperOption{wordwrap.BoxBox}},
		{400, 85, 16, 180, "goregular", "testdata/sample5.txt", "images/sample05.png", []wordwrap.WrapperOption{wordwrap.YOverflow(wordwrap.DescentOverflow)}},
		{200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample06.png", []wordwrap.WrapperOption{wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage))}},
		{200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample07.png", []wordwrap.WrapperOption{wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricAboveTheLine), wordwrap.BoxBox)}},
		{200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample08.png", []wordwrap.WrapperOption{wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricBelowTheLine), wordwrap.BoxBox)}},
		{200, 100, 16, 75, "goregular", "testdata/sample1.txt", "images/sample09.png", []wordwrap.WrapperOption{wordwrap.NewPageBreakBox(wordwrap.NewImageBox(chevronImage, wordwrap.ImageBoxMetricCenter(fontDrawer)), wordwrap.BoxBox)}},
		{200, 120, 16, 75, "goregular", "testdata/sample1.txt", "images/sample10.png", []wordwrap.WrapperOption{wordwrap.HorizontalCenterLines}},
		{200, 120, 16, 75, "goregular", "testdata/sample1.txt", "images/sample11.png", []wordwrap.WrapperOption{wordwrap.RightLines}},
		{200, 120, 16, 75, "goregular", "testdata/sample7.txt", "images/sample12.png", []wordwrap.WrapperOption{wordwrap.HorizontalCenterBlock}},
		{200, 120, 16, 75, "goregular", "testdata/sample7.txt", "images/sample13.png", []wordwrap.WrapperOption{wordwrap.RightBlock}},
		{200, 120, 16, 75, "goregular", "testdata/sample6.txt", "images/sample14.png", []wordwrap.WrapperOption{wordwrap.VerticalCenterBlock}},
		{200, 120, 16, 75, "goregular", "testdata/sample6.txt", "images/sample15.png", []wordwrap.WrapperOption{wordwrap.BottomBlock}},
	}

	for _, s := range samples {
		if err := SampleType1(s.w, s.h, s.size, s.dpi, s.font, s.src, s.dst, s.opts...); err != nil {
			return err
		}
	}
	return nil
}

func SampleType1(width, height int, fontsize, dpi float64, fontname, textsource, outfilename string, opts ...wordwrap.WrapperOption) error {
	log.Printf("Working on %s", outfilename)
	i := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(i, i.Bounds(), image.NewUniform(image.White), i.Bounds().Min, draw.Over)
	grf, err := getFontFace(fontname, fontsize, dpi)
	if err != nil {
		return fmt.Errorf("Error opening font %s: %w", fontname, err)
	}
	text, err := GetText(textsource)
	if err != nil {
		return fmt.Errorf("Text fetch error: %w", err)
	}
	sw := wordwrap.NewSimpleWrapper([]*wordwrap.Content{wordwrap.NewContent(text)}, grf, opts...)
	lines, _, err := sw.TextToRect(i.Bounds())
	if err != nil {
		return fmt.Errorf("Text wrap error: %w", err)
	}
	if err := sw.RenderLines(i, lines, i.Bounds().Min); err != nil {
		return fmt.Errorf("Text draw error: %w", err)
	}
	if err := SaveFile(i, outfilename); err != nil {
		return fmt.Errorf("Error with saving file: %w", err)
	}
	log.Printf("Done as %s", outfilename)
	return nil
}

func getFontFace(fontname string, fontsize float64, dpi float64) (font.Face, error) {
	gr, err := util.OpenFont(fontname)
	if err != nil {
		return nil, err
	}
	grf := util.GetFontFace(fontsize, dpi, gr)
	return grf, err
}
