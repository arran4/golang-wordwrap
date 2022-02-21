package wordwrap

import (
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"image"
	"reflect"
	"testing"
)

func TestSimpleBoxer_BoxNextWord(t *testing.T) {
	grf := FontFaceForTest(t)
	type args struct {
		fce   font.Face
		color image.Image
		text  []rune
	}
	tests := []struct {
		name             string
		args             args
		wantBoxString    string
		wantSimpleBox    bool
		wantLineBreakBox bool
		wantN            int
		wantErr          error
		wantNilBox       bool
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
			wantSimpleBox: true,
		},
		{
			name: "Empty string",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune(""),
			},
			wantBoxString: "",
			wantN:         len(""),
			wantSimpleBox: true,
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
			wantSimpleBox: true,
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
			wantSimpleBox: true,
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
			wantSimpleBox: true,
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
			wantSimpleBox: true,
		},
		{
			name: "Line break CRLF breaks words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("words\r\nhello"),
			},
			wantBoxString: "words",
			wantN:         len("words"),
			wantSimpleBox: true,
		},
		{
			name: "Line break LF breaks words",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("words\nhello"),
			},
			wantBoxString: "words",
			wantN:         len("words"),
			wantSimpleBox: true,
		},
		{
			name: "Line break CRLF breaks spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("    \r\nhello"),
			},
			wantBoxString: "    ",
			wantN:         len("    "),
			wantSimpleBox: true,
		},
		{
			name: "Line break LF breaks spaces",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("    \nhello"),
			},
			wantBoxString: "    ",
			wantN:         len("    "),
			wantSimpleBox: true,
		},
		{
			name: "Captures LF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\n"),
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\nhello"),
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\n "),
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\r\n"),
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\r\nword"),
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\r\n    "),
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\r\n\n"),
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFCRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune("\r\n\n\n"),
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Empty returns nil",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  []rune(""),
			},
			wantNilBox: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewSimpleBoxer(tt.args.text, &font.Drawer{
				Src:  tt.args.color,
				Face: tt.args.fce,
			})
			b, n, err := sb.Next()
			if tt.wantSimpleBox {
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
			} else if tt.wantNilBox {
				if b != nil {
					t.Errorf("BoxNextWord()[0] b = %s, wanted nil", reflect.TypeOf(b))
				}
			} else if tt.wantLineBreakBox {
				_, ok := b.(*LineBreakBox)
				if !ok {
					t.Errorf("BoxNextWord()[0] b = %s, wanted line break", reflect.TypeOf(b))
				}
			} else {
				t.Errorf("Unselected want for BoxNextWord()[0]")
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

func FontFaceForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 180, gr)
	return grf
}
