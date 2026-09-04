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
			name: "Image does not fit after preceding content",
			boxes: []Box{
				&ImageBox{
					I: image.NewRGBA(image.Rect(0, 0, 50, 100)),
					M: font.Metrics{Height: fixed.I(100)},
				},
				&ImageBox{
					I: image.NewRGBA(image.Rect(0, 0, 60, 100)), // 50+60 = 110 > 100
					M: font.Metrics{Height: fixed.I(100)},
				},
			},
			containerW:   100,
			expectedLine: 1,
			leftoverBox:  1,
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
