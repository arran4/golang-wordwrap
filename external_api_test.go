package wordwrap_test

import (
	"image"
	"testing"

	"github.com/arran4/golang-wordwrap"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// Ensure that an external wrapper can access Line stats
type statTrackingLine interface {
	wordwrap.Line
	SetStats(lineNumber int, pageNumber int, boxOffset int, currentPageBoxOffset int)
	GetStats() *wordwrap.LinePositionStats
}

func TestExternalLinePositionAndStatsAPI(t *testing.T) {
	content := wordwrap.NewContent("test")
	wrapper := wordwrap.NewSimpleWrapper([]*wordwrap.Content{content}, basicfont.Face7x13)

	lines, _, err := wrapper.TextToRect(image.Rect(0, 0, 100, 100))
	if err != nil {
		t.Fatalf("wrapper.TextToRect err: %v", err)
	}

	for _, l := range lines {
		// Can we call SetStats?
		sl, ok := l.(statTrackingLine)
		if !ok {
			t.Fatalf("Line doesn't implement statTrackingLine with SetStats() and GetStats()")
		}
		sl.SetStats(1, 2, 3, 4)
		stats := sl.GetStats()
		if stats == nil || stats.LineNumber != 1 {
			t.Fatalf("Stats did not match expected")
		}

		// Can we use HorizontalLinePosition?
		hlp, ok := l.(interface {
			GetHorizontalLinePosition() wordwrap.HorizontalLinePosition
		})
		if ok {
			pos := hlp.GetHorizontalLinePosition()
			if pos != wordwrap.LeftLines {
				t.Fatalf("Expected LeftLines as default")
			}
		} else {
			t.Fatalf("Line doesn't implement GetHorizontalLinePosition()")
		}

		slp, ok := l.(interface {
			SetHorizontalLinePosition(wordwrap.HorizontalLinePosition)
		})
		if ok {
			slp.SetHorizontalLinePosition(wordwrap.HorizontalCenterLines)
		} else {
			t.Fatalf("Line doesn't implement SetHorizontalLinePosition()")
		}
	}
}

func TestExternalFolderAPI(t *testing.T) {
	fd := &font.Drawer{
		Face: basicfont.Face7x13,
	}
	boxer := wordwrap.NewSimpleBoxer([]*wordwrap.Content{wordwrap.NewContent("test")}, fd)
	folder := wordwrap.NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)

	// Test GetPageBreakBox / SetPageBreakBox
	folder.SetPageBreakBox(&wordwrap.LineBreakBox{})
	pbb := folder.GetPageBreakBox()
	if pbb == nil {
		t.Fatalf("Expected page break box to be set")
	}

	// Test YOverflowMode
	mode := folder.YOverflowMode()
	if mode != wordwrap.StrictBorders {
		t.Fatalf("Expected StrictBorders as default")
	}

	// Test LastFontDrawer
	lfd := folder.LastFontDrawer()
	if lfd != nil {
		t.Fatalf("Expected nil LastFontDrawer initially for this test")
	}
}

// A custom implementation demonstrating external composition
type customExternalWrapper struct {
	folder *wordwrap.SimpleFolder
}

func TestExternalCustomWrapperBehavior(t *testing.T) {
	fd := &font.Drawer{
		Face: basicfont.Face7x13,
	}
	boxer := wordwrap.NewSimpleBoxer([]*wordwrap.Content{wordwrap.NewContent("hello world test")}, fd)
	// Apply descent overflow for this specific wrapper test using FolderOption cast
	opt := wordwrap.YOverflow(wordwrap.DescentOverflow).(wordwrap.FolderOption)
	wrapper := &customExternalWrapper{folder: wordwrap.NewSimpleFolder(boxer, image.Rect(0, 0, 50, 100), nil, opt)}

	line, err := wrapper.folder.Next(10)
	if err != nil {
		t.Fatalf("Failed to get line: %v", err)
	}
	if line == nil {
		t.Fatalf("Expected a line")
	}

	// Verify we can access YOverflowMode
	if wrapper.folder.YOverflowMode() != wordwrap.DescentOverflow {
		t.Fatalf("Expected DescentOverflow mode to be accessible and correct")
	}

	// Verify we can manipulate HorizontalLinePosition directly via the exposed interface
	if setter, ok := line.(interface {
		SetHorizontalLinePosition(wordwrap.HorizontalLinePosition)
	}); ok {
		setter.SetHorizontalLinePosition(wordwrap.RightLines)
	} else {
		t.Fatalf("Line cannot have its horizontal position set")
	}

	if getter, ok := line.(interface {
		GetHorizontalLinePosition() wordwrap.HorizontalLinePosition
	}); ok {
		if getter.GetHorizontalLinePosition() != wordwrap.RightLines {
			t.Fatalf("Expected horizontal position to be updated")
		}
	} else {
		t.Fatalf("Line cannot have its horizontal position retrieved")
	}

	// Make sure setting page break externally works and is readable
	wrapper.folder.SetPageBreakBox(&wordwrap.LineBreakBox{})
	if wrapper.folder.GetPageBreakBox() == nil {
		t.Fatalf("Expected PageBreakBox to be set")
	}
}

func TestLastFontDrawerIsTracked(t *testing.T) {
	fd1 := &font.Drawer{
		Face: basicfont.Face7x13,
	}

	content1 := wordwrap.NewContent("test1")
	// For testing, let's use a Boxer approach
	boxer := wordwrap.NewSimpleBoxer([]*wordwrap.Content{content1}, fd1)
	folder := wordwrap.NewSimpleFolder(boxer, image.Rect(0, 0, 100, 100), nil)

	// Initially nil
	if folder.LastFontDrawer() != nil {
		t.Fatalf("Expected LastFontDrawer to be nil before folding")
	}

	// Next parses and folds a line
	line, err := folder.Next(100)
	if err != nil {
		t.Fatalf("folder.Next failed: %v", err)
	}
	if line == nil {
		t.Fatalf("Expected a line to be returned")
	}

	// After folding, it should have the drawer used for the box
	if folder.LastFontDrawer() != fd1 {
		t.Fatalf("Expected LastFontDrawer to track the drawer after folding, got %v", folder.LastFontDrawer())
	}
}
