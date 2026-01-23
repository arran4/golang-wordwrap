package cli

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/golang-wordwrap/util"
	"golang.org/x/image/font"
)

// Unit defines a mechanism to get pixels
type Unit interface {
	Pixels(dpi float64) int
	String() string
}

type Pixels int

func (p Pixels) Pixels(dpi float64) int { return int(p) }
func (p Pixels) String() string         { return fmt.Sprintf("%dpx", int(p)) }

type Millimeters float64

func (mm Millimeters) Pixels(dpi float64) int { return int(float64(mm) / 25.4 * dpi) }
func (mm Millimeters) String() string         { return fmt.Sprintf("%.1fmm", float64(mm)) }

type Inches float64

func (in Inches) Pixels(dpi float64) int { return int(float64(in) * dpi) }
func (in Inches) String() string         { return fmt.Sprintf("%.1fin", float64(in)) }

// Constraint defines the bounds logic
type Constraint struct {
	// Min is the minimum size. If resolved pixels < Min, it is clamped to Min.
	Min Unit
	// Max is the maximum size. If resolved pixels > Max, it is clamped to Max.
	Max Unit
	// Unbounded overrides Max to infinity.
	Unbounded bool
	// Auto means the size should adapt to content.
	// For Width: adapts to max line width.
	// For Height: adapts to total used height.
	Auto bool
}

type Scenario struct {
	Name    string
	Width   Constraint
	Height  Constraint
	Content string
	DPI     float64
}

// PageLimits is a subcommand `wordwrap pagelimits`
func PageLimits() error {
	// Setup dependencies
	gr, err := util.OpenFont("goregular")
	if err != nil {
		return fmt.Errorf("failed to open font: %w", err)
	}
	fontRegular := util.GetFontFace(24, 96, gr)

	// Sample Text
	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
	longLorem := lorem + "\n\n" + lorem + "\n\n" + lorem

	// Define Scenarios
	scenarios := []Scenario{
		{
			Name:    "01_A4_Fixed_96DPI",
			Width:   Constraint{Min: Millimeters(210), Max: Millimeters(210)}, // Fixed A4 Width
			Height:  Constraint{Min: Millimeters(297), Max: Millimeters(297)}, // Fixed A4 Height
			Content: lorem,
			DPI:     96,
		},
		{
			Name:    "02_AutoWidth_MaxA4_96DPI",
			Width:   Constraint{Max: Millimeters(210), Auto: true}, // Shrink to fit, but max A4
			Height:  Constraint{Unbounded: true, Auto: true},       // Grow as needed
			Content: "Short line.",
			DPI:     96,
		},
		{
			Name:    "03_AutoWidth_MaxA4_Overflow_96DPI",
			Width:   Constraint{Max: Millimeters(210), Auto: true}, // Auto, but clipped/wrapped at A4
			Height:  Constraint{Unbounded: true, Auto: true},
			Content: longLorem, // Long content ensures wrapping at Max
			DPI:     96,
		},
		{
			Name:    "04_Fixed_300px_by_300px",
			Width:   Constraint{Min: Pixels(300), Max: Pixels(300)},
			Height:  Constraint{Min: Pixels(300), Max: Pixels(300)},
			Content: lorem,
			DPI:     96,
		},
		{
			Name:    "05_MinWidth_Measurement",
			Width:   Constraint{Min: Pixels(400), Max: Pixels(800), Auto: true}, // At least 400, max 800, fit content
			Height:  Constraint{Unbounded: true, Auto: true},
			Content: "Small.", // Should default to 400
			DPI:     96,
		},
		{
			Name:    "06_MinWidth_Measurement_LargeContent",
			Width:   Constraint{Min: Pixels(400), Max: Pixels(800), Auto: true},
			Height:  Constraint{Unbounded: true, Auto: true},
			Content: lorem, // Should expand beyond 400, maybe wrap at 800 or fit if smaller
			DPI:     96,
		},
		{
			Name:    "07_HighDPI_A4",
			Width:   Constraint{Min: Millimeters(210), Max: Millimeters(210)},
			Height:  Constraint{Min: Millimeters(297), Max: Millimeters(297)},
			Content: lorem,
			DPI:     300, // 300 DPI makes pixels larger implies image is larger
		},
	}

	for _, s := range scenarios {
		fmt.Printf("Running Scenario: %s\n", s.Name)
		if err := runScenario(s, fontRegular); err != nil {
			log.Printf("Error in scenario %s: %v", s.Name, err)
		}
	}
	return nil
}

func runScenario(s Scenario, grf font.Face) error {
	// Create factory for wrapper (content might need regeneration if stateful)
	createWrapper := func() *wordwrap.SimpleWrapper {
		return wordwrap.NewSimpleWrapper(s.Content, grf)
	}

	// Resolve Constraints
	minW, maxW, isUnboundedW := resolveConstraint(s.Width, s.DPI)
	if isUnboundedW {
		maxW = 100000 // Large finite basic limit for layout
	}

	targetWidth := maxW

	// Pass 1: Measure Width if Auto
	if s.Width.Auto {
		// Layout with 'Infinite' width to find natural width
		// Using a very large number that likely fits in Int
		infiniteW := 1000000
		w := createWrapper()
		lines, _, err := w.TextToRect(image.Rect(0, 0, infiniteW, 1000000))
		if err != nil {
			return fmt.Errorf("measure pass failed: %w", err)
		}

		maxLineWidth := 0
		for _, l := range lines {
			s := l.Size()
			// Check both rectangle bounds
			width := s.Dx()
			if width > maxLineWidth {
				maxLineWidth = width
			}
		}

		// Apply constraints
		targetWidth = maxLineWidth
		if targetWidth < minW {
			targetWidth = minW
		}
		if !isUnboundedW && targetWidth > maxW {
			targetWidth = maxW
		}
	}

	// Resolve Height Constraints
	minH, maxH, isUnboundedH := resolveConstraint(s.Height, s.DPI)
	if isUnboundedH {
		maxH = 100000 // Large layout limit
	}

	// Layout Pass (Final)
	layoutHeight := maxH
	if s.Height.Auto && isUnboundedH {
		layoutHeight = 100000 // Allow expansion
	}

	finalWrapper := createWrapper()
	lines, p, err := finalWrapper.TextToRect(image.Rect(0, 0, targetWidth, layoutHeight))
	if err != nil {
		return fmt.Errorf("layout pass failed: %w", err)
	}

	// Determine Final Canvas Height
	finalHeight := maxH
	if s.Height.Auto {
		// Height is determined by content
		contentHeight := p.Y
		finalHeight = contentHeight
		// Clamp
		if finalHeight < minH {
			finalHeight = minH
		}
		if !isUnboundedH && finalHeight > maxH {
			finalHeight = maxH
		}
	}

	// Render
	img := image.NewRGBA(image.Rect(0, 0, targetWidth, finalHeight))
	// Fill White
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw content
	if err := finalWrapper.RenderLines(img, lines, img.Bounds().Min); err != nil {
		return fmt.Errorf("render failed: %w", err)
	}

	// Save
	_ = os.MkdirAll(filepath.Join("cmd", "pagelimits"), 0755) // Original code saved here
	filename := filepath.Join("cmd", "pagelimits", s.Name+".png")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png failed: %w", err)
	}
	fmt.Printf("Generated %s (%dx%d)\n", filename, targetWidth, finalHeight)

	return nil
}

func resolveConstraint(c Constraint, dpi float64) (min, max int, unbounded bool) {
	minVal := 0
	if c.Min != nil {
		minVal = c.Min.Pixels(dpi)
	}
	maxVal := 0
	if c.Max != nil {
		maxVal = c.Max.Pixels(dpi)
	}
	// If uninitialized max or explicit unbounded
	if c.Unbounded || c.Max == nil {
		return minVal, 0, true
	}
	return minVal, maxVal, false
}
