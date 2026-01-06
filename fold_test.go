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

package wordwrap

import (
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"reflect"
	"testing"
)

func TestSimpleFolder(t *testing.T) {
	type WantedLine struct {
		words []string
		N     int
	}
	tests := []struct {
		name      string
		folder    Folder
		wantLines []*WantedLine
		wantErr   bool
	}{
		{
			name: "just word that fits",
			folder: NewSimpleFolder(&FixedWordWidthBoxer{
				text: []rune("word that fits"),
			}, image.Rect(0, 0, 6, 6), nil),
			wantLines: []*WantedLine{
				{
					words: []string{"word", " ", "that", " ", "fits"},
					N:     len("word that fits"),
				},
			},
			wantErr: false,
		},
		{
			name: "Empty",
			folder: NewSimpleFolder(&FixedWordWidthBoxer{
				text: []rune(""),
			}, image.Rect(0, 0, 2, 5), nil),
			wantLines: []*WantedLine{
				{
					words: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "word that folder over onto a new line",
			folder: NewSimpleFolder(&FixedWordWidthBoxer{
				text: []rune("word that folder over onto a new line"),
			}, image.Rect(0, 0, 6, 5), nil),
			wantLines: []*WantedLine{
				{
					words: []string{"word", " ", "that", " ", "folder"},
					N:     len("word that folder "),
				},
				{
					words: []string{"over", " ", "onto", " ", "a"},
					N:     len("over onto a "),
				},
				{
					words: []string{"new", " ", "line"},
					N:     len("new line"),
				},
			},
			wantErr: false,
		},
		{
			name: "eod is nil",
			folder: NewSimpleFolder(&FixedWordWidthBoxer{
				text: []rune("word that"),
			}, image.Rect(0, 0, 6, 5), nil),
			wantLines: []*WantedLine{
				{
					words: []string{"word", " ", "that"},
					N:     len("word that"),
				},
				nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, wantWords := range tt.wantLines {
				gotL, err := tt.folder.Next(7)
				if (err != nil) != tt.wantErr {
					t.Errorf("SimpleFolder() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				var gotWords []string
				switch l := gotL.(type) {
				case *SimpleLine:
					for _, b := range l.boxes {
						switch b := b.(type) {
						case *SimpleTextBox:
							gotWords = append(gotWords, b.Contents)
						}
					}
				}
				if wantWords == nil {
					if gotL != nil {
						t.Errorf("SimpleFolder() gotN = %v, expected no line", gotL)
					}
					return
				}
				if !reflect.DeepEqual(gotWords, wantWords.words) {
					t.Errorf("SimpleFolder() gotWords = %v, wantWords.words %v", gotWords, wantWords.words)
				}
			}
		})
	}
}

type FixedWordWidthBoxer struct {
	text []rune
	n    int
}

func (fwb *FixedWordWidthBoxer) Shift() Box {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) Unshift(b ...Box) {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) Pos() int {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) Push(box ...Box) {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) HasNext() bool {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) SetFontDrawer(face *font.Drawer) {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) FontDrawer() *font.Drawer {
	panic("implement me")
}

func (fwb *FixedWordWidthBoxer) Back(i int) {
	fwb.n -= i
}

func (fwb *FixedWordWidthBoxer) Next() (Box, int, error) {
	n, rs, rmode := SimpleBoxerGrab(fwb.text[fwb.n:])
	var b Box
	switch rmode {
	case RNIL:
		return nil, fwb.n, nil
	case RSimpleBox, RCRLF:
		t := string(rs)
		b = &SimpleTextBox{
			Contents: t,
			Bounds:   fixed.R(0, 0, 1, 1),
			Advance:  fixed.I(1),
		}
	default:
		return nil, fwb.n, fmt.Errorf("unknown rmode %d", rmode)
	}
	switch rmode {
	case RCRLF:
		b = &LineBreakBox{
			Box: b,
		}
	}
	fwb.n += n
	return b, fwb.n, nil
}
