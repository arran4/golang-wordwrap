package wordwrap

import (
	"image"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type manualBoxer struct {
	boxes []Box
}

func (m *manualBoxer) Next() (Box, int, error) {
	if len(m.boxes) == 0 {
		return nil, 0, nil
	}
	b := m.boxes[0]
	m.boxes = m.boxes[1:]
	return b, 0, nil
}

func (m *manualBoxer) Push(b ...Box)                   { m.boxes = append(b, m.boxes...) }
func (m *manualBoxer) Unshift(b ...Box)                { m.boxes = append(b, m.boxes...) }
func (m *manualBoxer) SetFontDrawer(face *font.Drawer) {}
func (m *manualBoxer) FontDrawer() *font.Drawer        { return nil }
func (m *manualBoxer) Init(face *font.Drawer)          {}
func (m *manualBoxer) HasNext() bool                   { return len(m.boxes) > 0 }
func (m *manualBoxer) Back(i int)                      {}
func (m *manualBoxer) Pos() int                        { return 0 }
func (m *manualBoxer) Shift() Box                      { return nil }
func (m *manualBoxer) Reset()                          {}

func TestSimpleFolder_ImageBox(t *testing.T) {
	tests := []struct {
		name         string
		boxes        []Box
		containerW   int
		expectedLine int // Number of boxes expected in the first line
		leftoverBox  int // Number of boxes expected left in the boxer
	}{
		{
			name: "Image wider than container on empty line",
			boxes: []Box{
				&ImageBox{
					I: image.NewRGBA(image.Rect(0, 0, 150, 100)),
					M: font.Metrics{Height: fixed.I(100)},
				},
			},
			containerW:   100,
			expectedLine: 1,
			leftoverBox:  0,
		},
		{
			name: "Exact width boundary behaviour",
			boxes: []Box{
				&ImageBox{
					I: image.NewRGBA(image.Rect(0, 0, 50, 100)),
					M: font.Metrics{Height: fixed.I(100)},
				},
				&ImageBox{
					I: image.NewRGBA(image.Rect(0, 0, 50, 100)), // 50+50 = 100 >= 100
					M: font.Metrics{Height: fixed.I(100)},
				},
			},
			containerW:   100,
			expectedLine: 1,
			leftoverBox:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boxer := &manualBoxer{boxes: tt.boxes}
			folder := NewSimpleFolder(boxer, image.Rect(0, 0, tt.containerW, 100), nil)

			line, err := folder.Next(100)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if line == nil {
				t.Fatalf("Line is nil, infinite loop expected")
			}
			l := line.(*SimpleLine)
			if len(l.boxes) != tt.expectedLine {
				t.Fatalf("Expected %d boxes in line, got %d", tt.expectedLine, len(l.boxes))
			}
			if len(boxer.boxes) != tt.leftoverBox {
				t.Fatalf("Expected %d boxes left in boxer, got %d", tt.leftoverBox, len(boxer.boxes))
			}

			// Verify repeated Next() calls make progress and don't loop on empty line
			if len(boxer.boxes) == 0 {
				next, err := folder.Next(100)
				if err != nil {
					t.Fatalf("Unexpected error on empty Next: %v", err)
				}
				if next != nil {
					t.Fatalf("Expected nil line when boxer is empty, got %v", next)
				}
			}
		})
	}
}

func TestSimpleFolder_ImageBox_PrecedingContent(t *testing.T) {
	b1 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 50, 100)),
		M: font.Metrics{Height: fixed.I(100)},
	}
	b2 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 60, 100)), // 50+60 = 110 > 100
		M: font.Metrics{Height: fixed.I(100)},
	}
	boxer := &manualBoxer{boxes: []Box{b1, b2}}
	folder := NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)

	// 1. Assert the first line contains only the 50px image.
	line1, err := folder.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	l1 := line1.(*SimpleLine)
	if len(l1.boxes) != 1 {
		t.Fatalf("Expected 1 box in line 1, got %d", len(l1.boxes))
	}
	if l1.boxes[0] != b1 {
		t.Fatalf("Expected first line to contain b1")
	}

	// 2. Call folder.Next(...) again and assert the second line contains the 60px image.
	line2, err := folder.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	l2 := line2.(*SimpleLine)
	if len(l2.boxes) != 1 {
		t.Fatalf("Expected 1 box in line 2, got %d", len(l2.boxes))
	}
	if l2.boxes[0] != b2 {
		t.Fatalf("Expected second line to contain b2")
	}

	// 3. Assert the boxer is then empty.
	if len(boxer.boxes) != 0 {
		t.Fatalf("Expected boxer to be empty, got %d", len(boxer.boxes))
	}

	// 4. Call folder.Next(...) once more and assert it returns nil.
	line3, err := folder.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if line3 != nil {
		t.Fatalf("Expected nil line when boxer is empty, got %v", line3)
	}
}

func TestSimpleFolder_ImageBox_Bug47(t *testing.T) {
	b1 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 150, 100)),
		M: font.Metrics{Height: fixed.I(100)},
	}
	b2 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 50, 100)),
		M: font.Metrics{Height: fixed.I(100)},
	}
	b3 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 51, 100)),
		M: font.Metrics{Height: fixed.I(100)},
	}
	b4 := &ImageBox{
		I: image.NewRGBA(image.Rect(0, 0, 49, 100)),
		M: font.Metrics{Height: fixed.I(100)},
	}

	// Test 1: Image wider than container on empty line
	boxer1 := &manualBoxer{boxes: []Box{b1}}
	folder1 := NewSimpleFolder(boxer1, image.Rect(0, 0, 100, 100), nil)
	line1, err := folder1.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if line1 == nil {
		t.Fatalf("Line is nil, infinite loop expected")
	}
	if len(line1.(*SimpleLine).boxes) != 1 {
		t.Fatalf("Expected 1 box in line, got %d", len(line1.(*SimpleLine).boxes))
	}

	// Test 2: Oversized image after preceding text/content
	boxer2 := &manualBoxer{boxes: []Box{b2, b1}}
	folder2 := NewSimpleFolder(boxer2, image.Rect(0, 0, 100, 100), nil)
	line2, err := folder2.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(line2.(*SimpleLine).boxes) != 1 {
		t.Fatalf("Expected 1 box in line, got %d", len(line2.(*SimpleLine).boxes))
	}
	if len(boxer2.boxes) != 1 {
		t.Fatalf("Expected 1 box left in boxer, got %d", len(boxer2.boxes))
	}

	line3, err := folder2.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(line3.(*SimpleLine).boxes) != 1 {
		t.Fatalf("Expected 1 box in line, got %d", len(line3.(*SimpleLine).boxes))
	}
	if line3.(*SimpleLine).boxes[0] != b1 {
		t.Fatalf("Expected second line to contain b1")
	}

	// Test 3: Repeated Next calls do not return unbounded empty lines
	line4, err := folder2.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if line4 != nil {
		t.Fatalf("Expected nil line when boxer is empty, got %v", line4)
	}

	// Test 4: Exact width boundary behaviour
	boxer3 := &manualBoxer{boxes: []Box{b2, b2}}
	folder3 := NewSimpleFolder(boxer3, image.Rect(0, 0, 100, 100), nil)
	line5, err := folder3.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(line5.(*SimpleLine).boxes) != 1 {
		t.Fatalf("Expected 1 box in line, got %d", len(line5.(*SimpleLine).boxes))
	}

	// Test 5: Slightly below boundary
	boxer4 := &manualBoxer{boxes: []Box{b2, b4}}
	folder4 := NewSimpleFolder(boxer4, image.Rect(0, 0, 100, 100), nil)
	line6, err := folder4.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(line6.(*SimpleLine).boxes) != 2 {
		t.Fatalf("Expected 2 boxes in line, got %d", len(line6.(*SimpleLine).boxes))
	}

	// Test 6: Slightly above boundary
	boxer5 := &manualBoxer{boxes: []Box{b2, b3}}
	folder5 := NewSimpleFolder(boxer5, image.Rect(0, 0, 100, 100), nil)
	line7, err := folder5.Next(100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(line7.(*SimpleLine).boxes) != 1 {
		t.Fatalf("Expected 1 box in line, got %d", len(line7.(*SimpleLine).boxes))
	}
}
