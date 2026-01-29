package wordwrap

import (
	"image"
	"reflect"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type ManualBoxer struct {
	boxes []Box
	n     int
	queue []Box
}

func (mb *ManualBoxer) Next() (Box, int, error) {
	if len(mb.queue) > 0 {
		b := mb.queue[0]
		mb.queue = mb.queue[1:]
		return b, 0, nil
	}
	if mb.n >= len(mb.boxes) {
		return nil, 0, nil
	}
	b := mb.boxes[mb.n]
	mb.n++
	return b, 1, nil
}

func (mb *ManualBoxer) SetFontDrawer(face *font.Drawer) {}
func (mb *ManualBoxer) FontDrawer() *font.Drawer          { return nil }
func (mb *ManualBoxer) Back(i int) {
	mb.n -= i
	if mb.n < 0 {
		mb.n = 0
	}
}
func (mb *ManualBoxer) HasNext() bool {
	return len(mb.queue) > 0 || mb.n < len(mb.boxes)
}
func (mb *ManualBoxer) Push(box ...Box) {
	mb.queue = append(mb.queue, box...)
}
func (mb *ManualBoxer) Pos() int { return mb.n }
func (mb *ManualBoxer) Unshift(box ...Box) {
	mb.queue = append(append(make([]Box, 0, len(box)+len(mb.queue)), box...), mb.queue...)
}
func (mb *ManualBoxer) Shift() Box {
	if len(mb.queue) > 0 {
		b := mb.queue[0]
		mb.queue = mb.queue[1:]
		return b
	}
	return nil
}
func (mb *ManualBoxer) Reset() {
	mb.n = 0
	mb.queue = nil
}

// Ensure interface compliance
var _ Boxer = (*ManualBoxer)(nil)

// Helper to create a simple text box for testing
func NewTestBox(s string, width int) *SimpleTextBox {
	return &SimpleTextBox{
		Contents: s,
		Advance:  fixedInt(width),
	}
}

func fixedInt(i int) fixed.Int26_6 {
	return fixed.I(i)
}

func TestFillLineBox(t *testing.T) {
	type WantedLine struct {
		words []string
	}

	tests := []struct {
		name           string
		boxes          []Box
		containerWidth int
		wantLines      []*WantedLine
	}{
		{
			name: "FillRestOfLine - fits",
			boxes: []Box{
				NewTestBox("A", 10),
				NewFillLineBox(NewTestBox("B", 10), FillRestOfLine),
				NewTestBox("C", 10),
			},
			containerWidth: 100,
			wantLines: []*WantedLine{
				{words: []string{"A", "B"}}, // B ends the line
				{words: []string{"C"}},
			},
		},
		{
			name: "FillRestOfLine - wraps",
			boxes: []Box{
				NewTestBox("A", 10),
				NewFillLineBox(NewTestBox("BigBox", 100), FillRestOfLine), // Overflows 10+100 > 50
				NewTestBox("C", 10),
			},
			containerWidth: 50,
			wantLines: []*WantedLine{
				{words: []string{"A"}},      // BigBox wraps
				{words: []string{"BigBox"}}, // BigBox is FillRestOfLine, so it ends this line
				{words: []string{"C"}},
			},
		},
		{
			name: "FillEntireLine - start",
			boxes: []Box{
				NewFillLineBox(NewTestBox("A", 10), FillEntireLine),
				NewTestBox("B", 10),
			},
			containerWidth: 50,
			wantLines: []*WantedLine{
				{words: []string{"A"}},
				{words: []string{"B"}},
			},
		},
		{
			name: "FillEntireLine - middle",
			boxes: []Box{
				NewTestBox("A", 10),
				NewFillLineBox(NewTestBox("B", 10), FillEntireLine),
				NewTestBox("C", 10),
			},
			containerWidth: 50,
			wantLines: []*WantedLine{
				{words: []string{"A"}}, // B forces break
				{words: []string{"B"}}, // B takes its own line
				{words: []string{"C"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boxer := &ManualBoxer{boxes: tt.boxes}
			folder := NewSimpleFolder(boxer, image.Rect(0, 0, tt.containerWidth, 100), nil)

			for i, wantLine := range tt.wantLines {
				gotL, err := folder.Next(100)
				if err != nil {
					t.Errorf("Next() error = %v", err)
					return
				}
				if gotL == nil {
					t.Errorf("Next() returned nil, expected line %d", i)
					return
				}

				var gotWords []string
				sl := gotL.(*SimpleLine)
				for _, b := range sl.boxes {
					// Helper to extract text
					txt := b.TextValue()
					gotWords = append(gotWords, txt)
				}

				if !reflect.DeepEqual(gotWords, wantLine.words) {
					t.Errorf("Line %d: got %v, want %v", i, gotWords, wantLine.words)
				}
			}

			// Ensure no more lines
			gotL, err := folder.Next(100)
			if err != nil {
				t.Errorf("Extra Next() error = %v", err)
			}
			if gotL != nil {
				t.Errorf("Extra line found: %v", gotL.TextValue())
			}
		})
	}
}
