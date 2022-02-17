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
		fce       font.Face
		feed      []rune
		container image.Rectangle
	}
	type WantedLine struct {
		words []string
		N     int
	}
	tests := []struct {
		name      string
		args      args
		wantLines []*WantedLine
		wantErr   bool
	}{
		{
			name: "just word that fits",
			args: args{
				boxer:     FixedWordWidthBoxer,
				fce:       nil,
				feed:      []rune("word that fits"),
				container: image.Rect(0, 0, 6, 6),
			},
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
			args: args{
				boxer:     FixedWordWidthBoxer,
				fce:       nil,
				feed:      []rune(""),
				container: image.Rect(0, 0, 2, 5),
			},
			wantLines: []*WantedLine{
				{
					words: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "word that folder over onto a new line",
			args: args{
				boxer:     FixedWordWidthBoxer,
				fce:       nil,
				feed:      []rune("word that folder over onto a new line"),
				container: image.Rect(0, 0, 6, 5),
			},
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
			args: args{
				boxer:     FixedWordWidthBoxer,
				fce:       nil,
				feed:      []rune("word that"),
				container: image.Rect(0, 0, 6, 5),
			},
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
			n := 0
			for _, wantWords := range tt.wantLines {
				gotL, gotN, err := SimpleFolder(tt.args.boxer, tt.args.fce, tt.args.feed[n:], tt.args.container)
				n += gotN
				if (err != nil) != tt.wantErr {
					t.Errorf("SimpleFolder() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				var gotWords []string
				switch l := gotL.(type) {
				case *SimpleLine:
					for _, b := range l.Boxes {
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
				if gotN != wantWords.N {
					t.Errorf("SimpleFolder() gotN = %v, wantWords.N %v", gotN, wantWords.N)
				}
			}
		})
	}
}

func FixedWordWidthBoxer(_ font.Face, _ image.Image, text []rune) (Box, int, error) {
	n, rs, rmode := SimpleBoxerGrab(text)
	switch rmode {
	case RNIL:
		return nil, n, nil
	case RCRLF:
		return &LineBreakBox{}, n, nil
	case RSimpleBox:
		t := string(rs)
		return &SimpleBox{
			Contents: t,
			Bounds:   fixed.R(0, 0, 1, 1),
		}, n, nil
	default:
		return nil, 0, fmt.Errorf("unknown rmode %d", rmode)
	}
}
