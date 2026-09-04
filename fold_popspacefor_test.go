package wordwrap

import (
	"fmt"
	"image"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func TestPopSpaceFor(t *testing.T) {
	// Simple mock box
	newBox := func(width int, isWs bool, text string) Box {
		return &mockBox{
			width: fixed.I(width),
			isWs:  isWs,
			text:  text,
		}
	}

	t.Run("displace multiple boxes and restore order", func(t *testing.T) {
		boxer := &FixedWordWidthBoxer{}
		folder := NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)
		line := folder.NewLine()

		line.Push(newBox(10, false, "A"), fixed.I(10))
		line.Push(newBox(10, false, "B"), fixed.I(10))
		line.Push(newBox(10, false, "C"), fixed.I(10))

		containerRect := image.Rect(0, 0, 80, 100)
		targetBox := newBox(70, false, "Target")

		c, err := line.PopSpaceFor(folder, containerRect, targetBox)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if c != 2 {
			t.Fatalf("expected 2 boxes popped, got %d", c)
		}

		// The remaining box should be A
		if len(line.boxes) != 2 {
			t.Fatalf("expected 2 boxes in line (A + Target), got %d", len(line.boxes))
		}
		if line.boxes[0].TextValue() != "A" {
			t.Fatalf("expected box A, got %s", line.boxes[0].TextValue())
		}
		if line.boxes[1].TextValue() != "Target" {
			t.Fatalf("expected Target box, got %s", line.boxes[1].TextValue())
		}

		// The boxer should have received B then C in order.
		if len(boxer.queue) != 2 {
			t.Fatalf("expected 2 boxes in boxer queue, got %d", len(boxer.queue))
		}
		if boxer.queue[0].TextValue() != "B" {
			t.Fatalf("expected B to be next, got %s", boxer.queue[0].TextValue())
		}
		if boxer.queue[1].TextValue() != "C" {
			t.Fatalf("expected C to be second, got %s", boxer.queue[1].TextValue())
		}
	})

	t.Run("error restoring boxes", func(t *testing.T) {
		boxer := &FixedWordWidthBoxer{}
		folder := NewSimpleFolder(boxer, image.Rect(0, 0, 80, 100), nil)
		line := folder.NewLine()

		line.Push(newBox(10, false, "A"), fixed.I(10))
		line.Push(newBox(10, false, "B"), fixed.I(10))

		containerRect := image.Rect(0, 0, 80, 100)
		targetBox := newBox(90, false, "Target")

		c, err := line.PopSpaceFor(folder, containerRect, targetBox)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if c != 0 {
			t.Fatalf("expected c=0, got %d", c)
		}

		// The boxer should have received A then B in order.
		if len(boxer.queue) != 2 {
			t.Fatalf("expected 2 boxes in boxer queue, got %d", len(boxer.queue))
		}
		if boxer.queue[0].TextValue() != "A" {
			t.Fatalf("expected A to be next, got %s", boxer.queue[0].TextValue())
		}
		if boxer.queue[1].TextValue() != "B" {
			t.Fatalf("expected B to be second, got %s", boxer.queue[1].TextValue())
		}
	})

	t.Run("PageBreakBox whitespace tracking", func(t *testing.T) {
		boxer := &FixedWordWidthBoxer{}
		folder := NewSimpleFolder(boxer, image.Rect(0, 0, 80, 100), nil)
		line := folder.NewLine()

		line.Push(newBox(10, false, "A"), fixed.I(10))

		// Push the whitespace box onto the line.
		wsBox := newBox(10, true, "WS")
		line.Push(wsBox, fixed.I(10))

		containerRect := image.Rect(0, 0, 80, 100)
		targetBox := NewPageBreak(newBox(65, false, "PB"))

		c, err := line.PopSpaceFor(folder, containerRect, targetBox)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if c != 0 {
			t.Fatalf("expected c=0, got %d", c)
		}

		if targetBox.ContainerBox == nil || targetBox.ContainerBox.TextValue() != "WS" {
			t.Fatalf("expected ContainerBox to be WS")
		}

		if len(boxer.queue) != 0 {
			t.Fatalf("expected boxer queue to be empty, got %d", len(boxer.queue))
		}
	})
}

// mockBox is a simple box for testing
type mockBox struct {
	width fixed.Int26_6
	isWs  bool
	text  string
}

func (m *mockBox) MinSize() (fixed.Int26_6, fixed.Int26_6)          { return m.width, 0 }
func (m *mockBox) MaxSize() (fixed.Int26_6, fixed.Int26_6)          { return m.width, 0 }
func (m *mockBox) AdvanceRect() fixed.Int26_6                       { return m.width }
func (m *mockBox) MetricsRect() font.Metrics                        { return font.Metrics{} }
func (m *mockBox) Whitespace() bool                                 { return m.isWs }
func (m *mockBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {}
func (m *mockBox) FontDrawer() *font.Drawer                         { return nil }
func (m *mockBox) Len() int                                         { return len(m.text) }
func (m *mockBox) TextValue() string                                { return m.text }

func BenchmarkPopSpaceFor(b *testing.B) {
	boxer := &FixedWordWidthBoxer{}
	folder := NewSimpleFolder(boxer, image.Rect(0, 0, 100, 10), nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := folder.NewLine()
		for j := 0; j < 1000; j++ {
			line.Push(&mockBox{width: fixed.I(10), text: "A"}, fixed.I(10))
		}

		// Target width is 950. We need to clear space to make line size + target size <= container.
		// Line currently has 1000 * 10 = 10000 width. Container is 1000.
		// target width 950, so we need line width <= 50.
		// So we will pop 995 boxes.
		targetBox := &mockBox{width: fixed.I(950), text: "Target"}
		b.StartTimer()

		_, err := line.PopSpaceFor(folder, image.Rect(0, 0, 1000, 10), targetBox)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}

		b.StopTimer()
		boxer.queue = nil // clear queue for next iteration
		b.StartTimer()
	}
}

// The baseline algorithm from previous PopSpaceFor, encapsulated here to verify parity
func popSpaceForBaseline(l *SimpleLine, sf *SimpleFolder, r image.Rectangle, box Box) (int, error) {
	ar := box.AdvanceRect()
	lastWs := false
	c := 0
	for r.Dx() < (l.size.Max.X - l.size.Min.X + ar).Ceil() {
		b := l.Pop()
		if b == nil {
			return 0, fmt.Errorf("no more boxes")
		}
		c++
		sf.boxer.Unshift(b)
		lastWs = b.Whitespace()
	}
	switch box := box.(type) {
	case *PageBreakBox:
		if lastWs {
			c--
			box.ContainerBox = sf.boxer.Shift()
		}
	}
	l.Push(box, ar)
	return c, nil
}

func TestPopSpaceForParity(t *testing.T) {
	newBox := func(width int, isWs bool, text string) Box {
		return &mockBox{width: fixed.I(width), isWs: isWs, text: text}
	}

	testCases := []struct {
		name          string
		containerRect image.Rectangle
		targetBox     func() Box
	}{
		{"success displacement", image.Rect(0, 0, 80, 100), func() Box { return newBox(70, false, "Target") }},
		{"error displacement", image.Rect(0, 0, 80, 100), func() Box { return newBox(90, false, "Target") }},
		{"pagebreak whitespace tracking", image.Rect(0, 0, 80, 100), func() Box { return NewPageBreak(newBox(65, false, "PB")) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup for Baseline
			boxerBase := &FixedWordWidthBoxer{}
			folderBase := NewSimpleFolder(boxerBase, image.Rect(0, 0, 100, 100), nil)
			lineBase := folderBase.NewLine()
			lineBase.Push(newBox(10, false, "A"), fixed.I(10))
			lineBase.Push(newBox(10, true, "WS"), fixed.I(10))
			lineBase.Push(newBox(10, false, "C"), fixed.I(10))
			targetBase := tc.targetBox()

			// Setup for Optimized
			boxerOpt := &FixedWordWidthBoxer{}
			folderOpt := NewSimpleFolder(boxerOpt, image.Rect(0, 0, 100, 100), nil)
			lineOpt := folderOpt.NewLine()
			lineOpt.Push(newBox(10, false, "A"), fixed.I(10))
			lineOpt.Push(newBox(10, true, "WS"), fixed.I(10))
			lineOpt.Push(newBox(10, false, "C"), fixed.I(10))
			targetOpt := tc.targetBox()

			// Execute Baseline
			cBase, errBase := popSpaceForBaseline(lineBase, folderBase, tc.containerRect, targetBase)

			// Execute Optimized
			cOpt, errOpt := lineOpt.PopSpaceFor(folderOpt, tc.containerRect, targetOpt)

			// Assertions
			if (errBase != nil && errOpt == nil) || (errBase == nil && errOpt != nil) {
				t.Fatalf("error parity mismatch: base=%v, opt=%v", errBase, errOpt)
			}

			if cBase != cOpt {
				t.Fatalf("count parity mismatch: base=%d, opt=%d", cBase, cOpt)
			}

			if len(boxerBase.queue) != len(boxerOpt.queue) {
				t.Fatalf("queue len mismatch: base=%d, opt=%d", len(boxerBase.queue), len(boxerOpt.queue))
			}

			for i, b := range boxerBase.queue {
				if b.TextValue() != boxerOpt.queue[i].TextValue() {
					t.Fatalf("queue parity mismatch at index %d: base=%s, opt=%s", i, b.TextValue(), boxerOpt.queue[i].TextValue())
				}
			}

			if len(lineBase.boxes) != len(lineOpt.boxes) {
				t.Fatalf("line boxes len mismatch: base=%d, opt=%d", len(lineBase.boxes), len(lineOpt.boxes))
			}

			for i, b := range lineBase.boxes {
				if b.TextValue() != lineOpt.boxes[i].TextValue() {
					t.Fatalf("line boxes parity mismatch at index %d: base=%s, opt=%s", i, b.TextValue(), lineOpt.boxes[i].TextValue())
				}
			}

			if pbBase, ok := targetBase.(*PageBreakBox); ok {
				pbOpt := targetOpt.(*PageBreakBox)

				var cBaseText, cOptText string
				if pbBase.ContainerBox != nil {
					cBaseText = pbBase.ContainerBox.TextValue()
				}
				if pbOpt.ContainerBox != nil {
					cOptText = pbOpt.ContainerBox.TextValue()
				}

				if cBaseText != cOptText {
					t.Fatalf("PageBreakBox ContainerBox parity mismatch: base=%s, opt=%s", cBaseText, cOptText)
				}
			}
		})
	}
}
