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
	"golang.org/x/image/math/fixed"
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
	if err := SampleGameMenu(); err != nil {
		return err
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

func SampleGameMenu() error {
	width := 800
	height := 600
	outfilename := "images/gamemenu/menu.png"
	log.Printf("Working on %s", outfilename)
	i := image.NewRGBA(image.Rect(0, 0, width, height))
	// Dark background
	draw.Draw(i, i.Bounds(), image.NewUniform(image.Black), i.Bounds().Min, draw.Over)

	// Font setup
	grf, err := getFontFace("goregular", 24, 75)
	if err != nil {
		return fmt.Errorf("Error opening font: %w", err)
	}
	drawer := &font.Drawer{
		Src:  image.NewUniform(image.White),
		Face: grf,
	}

	// Helper to create a menu item line
	menuItem := func(text string, highlighted bool) wordwrap.Box {
		// Basic text box
		var b wordwrap.Box

		if highlighted {
			// Highlighted: Black text
			highlightDrawer := &font.Drawer{
				Src:  image.NewUniform(image.Black),
				Face: grf,
			}
			tb, _ := wordwrap.NewSimpleTextBox(highlightDrawer, text)
			b = tb
		} else {
			// Normal: White text
			tb, _ := wordwrap.NewSimpleTextBox(drawer, text)
			b = tb
		}

		// 1. Text Padding
		textPadding := fixed.R(20, 10, 20, 10)
		b = wordwrap.NewDecorationBox(b, textPadding, fixed.R(0,0,0,0), nil, wordwrap.BgPositioningZeroed)

		// 2. Background (Inner)
		bgColor := image.NewUniform(image.Black)
		if highlighted {
			bgColor = image.NewUniform(image.White)
		}
		b = &wordwrap.BackgroundBox{
			Box:        b,
			Background: bgColor,
		}

		// 3. Border Thickness (via Padding) + Border Color (via Background)
		// Border Thickness: 4px
		borderThickness := fixed.R(4, 4, 4, 4)

		// Border Color: Dark Grey for normal, Gold for highlighted?
		// Keeping simple: White border for all
		borderColor := image.NewUniform(image.White)

		// To make the border visible, we wrap the current box (which is content+bg)
		// in a DecorationBox that adds padding (the border thickness),
		// and then wrap THAT in a BackgroundBox that fills that padding with the border color.

		// Add padding for border thickness
		b = wordwrap.NewDecorationBox(b, borderThickness, fixed.R(0,0,0,0), nil, wordwrap.BgPositioningZeroed)

		// Add background for border color
		b = &wordwrap.BackgroundBox{
			Box: b,
			Background: borderColor,
		}

		// 4. Outer Margin (spacing between items)
		margin := fixed.R(0, 10, 0, 10)
		b = wordwrap.NewDecorationBox(b, fixed.R(0,0,0,0), margin, nil, wordwrap.BgPositioningZeroed)

		// 5. Alignment and Fill Line
		b = &wordwrap.AlignedBox{
			Box: b,
			Alignment: wordwrap.AlignMiddle,
		}

		// Wrap in FillLineBox to consume the line (forcing new line behavior)
		flb := wordwrap.NewFillLineBox(b, wordwrap.FillEntireLine)
		return flb
	}

	boxes := []wordwrap.Box{
		// Title manually
		wordwrap.NewFillLineBox(
			&wordwrap.AlignedBox{
				Box: func() wordwrap.Box {
					b, _ := wordwrap.NewSimpleTextBox(drawer, "GAME MENU")
					return b
				}(),
				Alignment: wordwrap.AlignMiddle,
			},
			wordwrap.FillEntireLine,
		),
		// Spacing - Image box with height
		wordwrap.NewFillLineBox(wordwrap.NewImageBox(image.NewRGBA(image.Rect(0, 0, 1, 50))), wordwrap.FillEntireLine), // Spacer

		menuItem("NEW GAME", true), // Highlighted
		menuItem("LOAD GAME", false),
		menuItem("OPTIONS", false),
		menuItem("EXIT", false),
	}

	// Manual Boxer using refinedBoxer struct
	myBoxer := &refinedBoxer{
		boxes: boxes,
	}

	// Folder with HorizontalCenterLines to center the text within the filled lines
	folder := wordwrap.NewSimpleFolder(myBoxer, i.Bounds(), drawer, wordwrap.HorizontalCenterLines)

	// Layout
	y := 0
	for {
		line, err := folder.Next(height - y)
		if err != nil {
			return fmt.Errorf("layout error: %w", err)
		}
		if line == nil {
			break
		}

		// Draw
		// DrawLine usually draws at (0,0) of the provided image.
		// We need to provide a subimage starting at y.
		r := image.Rect(0, y, width, y+line.Size().Dy())
		if r.Dy() > 0 {
			subI := i.SubImage(r).(wordwrap.Image)
			// DrawLine expects an Image interface which typically includes SubImage.
			// NewRGBA implements Image.
			if err := line.DrawLine(subI); err != nil {
				return fmt.Errorf("draw line error: %w", err)
			}
		}
		y += line.Size().Dy()
	}

	if err := SaveFile(i, outfilename); err != nil {
		return fmt.Errorf("Error with saving file: %w", err)
	}
	log.Printf("Done as %s", outfilename)
	return nil
}

type refinedBoxer struct {
	boxes []wordwrap.Box
	n     int
	queue []wordwrap.Box
}

func (rb *refinedBoxer) Next() (wordwrap.Box, int, error) {
	if len(rb.queue) > 0 {
		b := rb.queue[0]
		rb.queue = rb.queue[1:]
		return b, 0, nil
	}
	if rb.n >= len(rb.boxes) {
		return nil, 0, nil
	}
	b := rb.boxes[rb.n]
	rb.n++
	return b, 1, nil
}
func (rb *refinedBoxer) SetFontDrawer(face *font.Drawer) {}
func (rb *refinedBoxer) FontDrawer() *font.Drawer          { return nil }
func (rb *refinedBoxer) Back(i int)                      { rb.n -= i; if rb.n < 0 { rb.n = 0 } }
func (rb *refinedBoxer) HasNext() bool                   { return len(rb.queue) > 0 || rb.n < len(rb.boxes) }
func (rb *refinedBoxer) Push(box ...wordwrap.Box)        { rb.queue = append(rb.queue, box...) }
func (rb *refinedBoxer) Pos() int                        { return rb.n }
func (rb *refinedBoxer) Unshift(box ...wordwrap.Box) {
	rb.queue = append(append(make([]wordwrap.Box, 0, len(box)+len(rb.queue)), box...), rb.queue...)
}
func (rb *refinedBoxer) Shift() wordwrap.Box {
	if len(rb.queue) > 0 {
		b := rb.queue[0]
		rb.queue = rb.queue[1:]
		return b
	}
	return nil
}
func (rb *refinedBoxer) Reset() { rb.n = 0; rb.queue = nil }
