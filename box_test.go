package wordwrap

import (
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"image"
	"log"
	"reflect"
	"testing"
)

func TestSimpleBoxer_BoxNextWord(t *testing.T) {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		log.Panicf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 180, gr)
	type args struct {
		fce   font.Face
		color image.Image
		text  []rune
	}
	tests := []struct {
		name          string
		args          args
		wantBoxString string
		wantN         int
		wantErr       error
	}{
		{
			name: "One word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("OnEWorD"),
			},
			wantBoxString: "OnEWorD",
			wantN:         len("OnEWorD"),
		},
		{
			name: "Multiple spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("   "),
			},
			wantBoxString: "   ",
			wantN:         len("   "),
		},
		{
			name: "Two words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("two words"),
			},
			wantBoxString: "two",
			wantN:         len("two"),
		},
		{
			name: "Two words multiple spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("two    words"),
			},
			wantBoxString: "two",
			wantN:         len("two"),
		},
		{
			name: "multiple spaces then one word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("    words"),
			},
			wantBoxString: "    ",
			wantN:         len("    "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := SimpleBoxer{}
			b, n, err := si.BoxNextWord(tt.args.fce, tt.args.color, tt.args.text)
			sb, ok := b.(*SimpleBox)
			if ok {
				if !reflect.DeepEqual(sb.Contents, tt.wantBoxString) {
					t.Errorf("BoxNextWord()[0].Contents b = %v, wantBoxString %v", sb.Contents, tt.wantBoxString)
				}
			} else {
				if len(tt.wantBoxString) > 0 {
					t.Errorf("BoxNextWord()[0].Contents b = %v, wantBoxString %v", b, tt.wantBoxString)
				}
			}
			if n != tt.wantN {
				t.Errorf("BoxNextWord() n = %v, wantN %v", n, tt.wantN)
			}
			if err != tt.wantErr {
				t.Errorf("BoxNextWord().error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
