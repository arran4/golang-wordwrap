// Copyright 2024 arran4
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
