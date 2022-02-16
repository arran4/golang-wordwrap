package util

import (
	"errors"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

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
