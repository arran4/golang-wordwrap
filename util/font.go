package util

import (
	"errors"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomediumitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/gofont/gosmallcaps"
	"golang.org/x/image/font/gofont/gosmallcapsitalic"
	"golang.org/x/image/font/inconsolata"
)

func GetFontFace(fontsize float64, dpi float64, gr *truetype.Font) font.Face {
	return truetype.NewFace(gr, &truetype.Options{
		Size: fontsize,
		DPI:  dpi,
	})
}

// OpenFont loads a TrueType font by name.
// Note: This currently only supports TrueType fonts provided by the gofont packages.
// Bitmap fonts like basicfont or inconsolata cannot be returned as *truetype.Font.
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

// GetFace returns a font.Face for the given font name.
// This supports both TrueType fonts (scaled to the given size and DPI)
// and bitmap fonts (which ignore size and DPI).
func GetFace(name string, fontsize float64, dpi float64) (font.Face, error) {
	// First check for bitmap fonts
	switch name {
	case "basicfont":
		return basicfont.Face7x13, nil
	case "inconsolata-bold":
		return inconsolata.Bold8x16, nil
	case "inconsolata-regular":
		return inconsolata.Regular8x16, nil
	}

	// Try to load as TrueType
	gr, err := OpenFont(name)
	if err != nil {
		return nil, err
	}
	return GetFontFace(fontsize, dpi, gr), nil
}

func FontByName(name string) ([]byte, error) {
	switch name {
	case "goregular":
		return goregular.TTF, nil
	case "gobold":
		return gobold.TTF, nil
	case "gobolditalic":
		return gobolditalic.TTF, nil
	case "goitalic":
		return goitalic.TTF, nil
	case "gomedium":
		return gomedium.TTF, nil
	case "gomediumitalic":
		return gomediumitalic.TTF, nil
	case "gomono":
		return gomono.TTF, nil
	case "gomonobold":
		return gomonobold.TTF, nil
	case "gomonobolditalic":
		return gomonobolditalic.TTF, nil
	case "gomonoitalic":
		return gomonoitalic.TTF, nil
	case "gosmallcaps":
		return gosmallcaps.TTF, nil
	case "gosmallcapsitalic":
		return gosmallcapsitalic.TTF, nil
	}
	return nil, errors.New("font not found")
}
