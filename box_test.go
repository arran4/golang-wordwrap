package wordwrap

import (
	"image"
	"image/color"
	"reflect"
	"testing"

	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

func TestSimpleBoxer_BoxNextWord(t *testing.T) {
	grf := FontFace16DPI180ForTest(t)
	type args struct {
		fce   font.Face
		color image.Image
		text  string
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
				text:  "OnEWorD",
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
				text:  "",
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
				text:  "   ",
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
				text:  "two words",
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
				text:  "two    words",
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
				text:  "    words",
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
				text:  "words\r\nhello",
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
				text:  "words\nhello",
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
				text:  "    \r\nhello",
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
				text:  "    \nhello",
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
				text:  "\n",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\nhello",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures LF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\n ",
			},
			wantN:            len("\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not word",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\nword",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not space",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n    ",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Captures CRLF and not CRLFCRLF",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "\r\n\n\n",
			},
			wantN:            len("\r\n"),
			wantLineBreakBox: true,
		},
		{
			name: "Empty returns nil",
			args: args{
				fce:   grf,
				color: image.NewUniform(colornames.Black),
				text:  "",
			},
			wantNilBox: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewSimpleBoxer([]rune(tt.args.text), &font.Drawer{
				Src:  tt.args.color,
				Face: tt.args.fce,
			})
			b, n, err := sb.Next()
			if tt.wantSimpleBox {
				sb, ok := b.(*SimpleTextBox)
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

func FontFace16DPI180ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 180, gr)
	return grf
}

func FontFace16DPI75ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(16, 75, gr)
	return grf
}

func FontFace24DPI180ForTest(t *testing.T) font.Face {
	gr, err := util.OpenFont("goregular")
	if err != nil {
		t.Errorf("Error opening font %s: %s", "goregular", err)
	}
	grf := util.GetFontFace(24, 180, gr)
	return grf
}

func TestSimpleBoxer_Reset(t *testing.T) {
	text := "Hello World"
	fontFace := FontFace16DPI180ForTest(t)
	drawer := &font.Drawer{Face: fontFace, Src: image.NewUniform(color.Black)}

	// contents := []*Content{{text: text}}
	boxer := NewSimpleBoxer([]rune(text), drawer)

	// Consume some boxes
	_, _, err := boxer.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	// Reset
	boxer.Reset()

	// Should be at start
	// "Hello" (SimpleTextBox)
	b, _, err := boxer.Next()
	if err != nil {
		t.Fatalf("Next after Reset failed: %v", err)
	}

	sb, ok := b.(*SimpleTextBox)
	if !ok {
		t.Fatalf("Expected SimpleTextBox, got %T", b)
	}

	// Depending on boxer logic, first box of "Hello World" is "Hello"
	if sb.Contents != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", sb.Contents)
	}
}
