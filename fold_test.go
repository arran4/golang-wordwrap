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
	type args struct {
		boxer     Boxer
		container image.Rectangle
		Options   []FolderOption
	}
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
			}, image.Rect(0, 0, 6, 6)),
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
			}, image.Rect(0, 0, 2, 5)),
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
			}, image.Rect(0, 0, 6, 5)),
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
			}, image.Rect(0, 0, 6, 5)),
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
				gotL, err := tt.folder.Next()
				if (err != nil) != tt.wantErr {
					t.Errorf("SimpleFolder() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				var gotWords []string
				switch l := gotL.(type) {
				case *SimpleLine:
					for _, b := range l.boxes {
						switch b := b.(type) {
						case *SimpleBox:
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
	case RCRLF:
		b = &LineBreakBox{}
	case RSimpleBox:
		t := string(rs)
		b = &SimpleBox{
			Contents: t,
			Bounds:   fixed.R(0, 0, 1, 1),
			Advance:  fixed.I(1),
		}
	default:
		return nil, fwb.n, fmt.Errorf("unknown rmode %d", rmode)
	}
	fwb.n += n
	return b, fwb.n, nil
}
