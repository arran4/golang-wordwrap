package wordwrap

import (
	"image"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type fillManualBoxer struct {
	boxes []Box
}

func (m *fillManualBoxer) Next() (Box, int, error) {
	if len(m.boxes) == 0 {
		return nil, 0, nil
	}
	b := m.boxes[0]
	m.boxes = m.boxes[1:]
	return b, 0, nil
}

func (m *fillManualBoxer) Push(b ...Box)                   { m.boxes = append(b, m.boxes...) }
func (m *fillManualBoxer) Unshift(b ...Box)                { m.boxes = append(b, m.boxes...) }
func (m *fillManualBoxer) SetFontDrawer(face *font.Drawer) {}
func (m *fillManualBoxer) FontDrawer() *font.Drawer        { return nil }
func (m *fillManualBoxer) Init(face *font.Drawer)          {}
func (m *fillManualBoxer) HasNext() bool                   { return len(m.boxes) > 0 }
func (m *fillManualBoxer) Back(i int)                      {}
func (m *fillManualBoxer) Pos() int                        { return 0 }
func (m *fillManualBoxer) Shift() Box                      { return nil }
func (m *fillManualBoxer) Reset()                          {}

type dummyBox struct {
	width  int
	height int
}

func (d *dummyBox) AdvanceRect() fixed.Int26_6 { return fixed.I(d.width) }
func (d *dummyBox) MetricsRect() font.Metrics {
	return font.Metrics{Height: fixed.I(d.height), Ascent: fixed.I(d.height)}
}
func (d *dummyBox) Whitespace() bool                                 { return false }
func (d *dummyBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {}
func (d *dummyBox) FontDrawer() *font.Drawer                         { return nil }
func (d *dummyBox) Len() int                                         { return 0 }
func (d *dummyBox) TextValue() string                                { return "" }
func (d *dummyBox) MinSize() (fixed.Int26_6, fixed.Int26_6)          { return 0, 0 }
func (d *dummyBox) MaxSize() (fixed.Int26_6, fixed.Int26_6)          { return 0, 0 }

func TestFoldFillModes(t *testing.T) {
	tests := []struct {
		name           string
		containerW     int
		boxes          []Box
		expectedLines  int
		expectedWidths []int
	}{
		{
			name:       "1. FillRestOfLine after normal content",
			containerW: 100,
			boxes: []Box{
				&dummyBox{width: 30, height: 10},
				&FillLineBox{Mode: FillRestOfLine, Box: &dummyBox{width: 10, height: 10}},
			},
			expectedLines:  1,
			expectedWidths: []int{100},
		},
		{
			name:       "2. FillEntireLine after normal content",
			containerW: 100,
			boxes: []Box{
				&dummyBox{width: 30, height: 10},
				&FillLineBox{Mode: FillEntireLine, Box: &dummyBox{width: 10, height: 10}},
			},
			expectedLines:  2,
			expectedWidths: []int{30, 100},
		},
		{
			name:       "3. FillEntireLine as the first item",
			containerW: 100,
			boxes: []Box{
				&FillLineBox{Mode: FillEntireLine, Box: &dummyBox{width: 10, height: 10}},
			},
			expectedLines:  1,
			expectedWidths: []int{100},
		},
		{
			name:       "4. oversized FillRestOfLine",
			containerW: 100,
			boxes: []Box{
				&dummyBox{width: 30, height: 10},
				&FillLineBox{Mode: FillRestOfLine, Box: &dummyBox{width: 150, height: 10}},
			},
			expectedLines:  2,
			expectedWidths: []int{30, 150},
		},
		{
			name:       "5. oversized FillEntireLine",
			containerW: 100,
			boxes: []Box{
				&FillLineBox{Mode: FillEntireLine, Box: &dummyBox{width: 150, height: 10}},
			},
			expectedLines:  1,
			expectedWidths: []int{150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boxer := &fillManualBoxer{boxes: tt.boxes}
			folder := NewSimpleFolder(boxer, image.Rect(0, 0, tt.containerW, 100), nil)

			var lines []Line
			for {
				line, err := folder.Next(100)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if line == nil {
					break
				}
				lines = append(lines, line)
			}

			if len(lines) != tt.expectedLines {
				t.Fatalf("Expected %d lines, got %d", tt.expectedLines, len(lines))
			}

			for i, line := range lines {
				sl, ok := line.(*SimpleLine)
				if !ok {
					t.Fatalf("Line %d is not a SimpleLine", i)
				}
				sz := sl.Size()
				w := sz.Dx()
				if w != tt.expectedWidths[i] {
					t.Errorf("Line %d expected width %d, got %d", i, tt.expectedWidths[i], w)
				}
			}
		})
	}
}

func TestFillLineBoxDrawingGeometry(t *testing.T) {
	bgImg := image.NewUniform(image.Black)
	decoratedBox := NewDecorationBox(
		&dummyBox{width: 20, height: 10},
		fixed.R(0, 0, 0, 0),
		fixed.R(0, 0, 0, 0),
		bgImg,
		BgPositioningZeroed,
	)

	fillBox := &FillLineBox{Mode: FillEntireLine, Box: decoratedBox}

	boxer := &fillManualBoxer{boxes: []Box{fillBox}}
	folder := NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)

	line, err := folder.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if line == nil {
		t.Fatalf("Expected line, got nil")
	}

	line.setStats(0, 0, 0, 0)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	err = line.DrawLine(img)
	if err != nil {
		t.Fatalf("DrawLine error: %v", err)
	}

	_, _, _, a := img.At(99, 5).RGBA()
	if a == 0 {
		t.Errorf("Expected pixel at x=99 to be drawn by DecorationBox, but it was transparent")
	}
}

func TestFillLineBoxContentFollowing(t *testing.T) {
	testCases := []struct {
		name string
		mode FillMode
	}{
		{"FillRestOfLine", FillRestOfLine},
		{"FillEntireLine", FillEntireLine},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			boxes := []Box{
				&FillLineBox{Mode: tc.mode, Box: &dummyBox{width: 10, height: 10}},
				&dummyBox{width: 30, height: 10},
			}
			boxer := &fillManualBoxer{boxes: boxes}
			folder := NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)

			var lines []Line
			for {
				line, err := folder.Next(100)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if line == nil {
					break
				}
				lines = append(lines, line)
			}

			if len(lines) != 2 {
				t.Fatalf("Expected 2 lines, got %d", len(lines))
			}

			// First line should ONLY contain the FillLineBox
			boxesOnLine1 := lines[0].Boxes()
			if len(boxesOnLine1) != 1 {
				t.Errorf("Expected first line to contain 1 box, got %d", len(boxesOnLine1))
			}
			if _, ok := boxesOnLine1[0].(*FillLineBox); !ok {
				t.Errorf("Expected first box on first line to be FillLineBox, got %T", boxesOnLine1[0])
			}

			// Second line should contain the following content
			boxesOnLine2 := lines[1].Boxes()
			if len(boxesOnLine2) != 1 {
				t.Errorf("Expected second line to contain 1 box, got %d", len(boxesOnLine2))
			}
			if _, ok := boxesOnLine2[0].(*dummyBox); !ok {
				t.Errorf("Expected first box on second line to be dummyBox, got %T", boxesOnLine2[0])
			}
		})
	}
}
