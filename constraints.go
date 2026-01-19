package wordwrap

import (
	"fmt"
	"image"
	"image/color"
)

// SizeFunction determines a size based on the measured content size
type SizeFunction func(contentSize int) int

// SpecOption configures the TextToSpecs layout
type SpecOption interface {
	ApplySpec(config *SpecConfig)
}

// SpecConfig holds configuration for TextToSpecs
type SpecConfig struct {
	WidthFn        SizeFunction
	HeightFn       SizeFunction
	Margin         SpecMargin
	PageBackground color.Color
}

type SpecMargin struct {
	Top, Right, Bottom, Left int
	Color                    color.Color
}

// WidthOption configures the width constraint
type WidthOption SizeFunction

func (f WidthOption) ApplySpec(c *SpecConfig) { c.WidthFn = SizeFunction(f) }

// HeightOption configures the height constraint
type HeightOption SizeFunction

func (f HeightOption) ApplySpec(c *SpecConfig) { c.HeightFn = SizeFunction(f) }

// Width sets the width constraint
func Width(f SizeFunction) WidthOption { return WidthOption(f) }

// Height sets the height constraint
func Height(f SizeFunction) HeightOption { return HeightOption(f) }

// PageMargin sets the margin option
type PageMarginOption struct {
	Margin int
	Color  color.Color
}

func (o PageMarginOption) ApplySpec(c *SpecConfig) {
	c.Margin.Top = o.Margin
	c.Margin.Right = o.Margin
	c.Margin.Bottom = o.Margin
	c.Margin.Left = o.Margin
	c.Margin.Color = o.Color
}

// Padding sets the margin option (Alias for PageMargin)
func Padding(margin int, c color.Color) PageMarginOption {
	return PageMarginOption{Margin: margin, Color: c}
}

// PageBackgroundOption sets the page background color
type PageBackgroundOption struct {
	Color color.Color
}

func (o PageBackgroundOption) ApplySpec(c *SpecConfig) {
	c.PageBackground = o.Color
}

func PageBackground(c color.Color) PageBackgroundOption {
	return PageBackgroundOption{Color: c}
}

// Helper Functions

// Auto returns the content size (identity). Same as Unbounded in this context.
func Auto() SizeFunction { return func(n int) int { return n } }

// Unbounded returns the content size.
func Unbounded() SizeFunction { return func(n int) int { return n } }

// Fixed returns a fixed size.
func Fixed(n int) SizeFunction { return func(_ int) int { return n } }

// Min returns the minimum of two size functions.
func Min(a, b SizeFunction) SizeFunction {
	return func(n int) int {
		v1 := a(n)
		v2 := b(n)
		if v1 < v2 {
			return v1
		}
		return v2
	}
}

// Max returns the maximum of two size functions.
func Max(a, b SizeFunction) SizeFunction {
	return func(n int) int {
		v1 := a(n)
		v2 := b(n)
		if v1 > v2 {
			return v1
		}
		return v2
	}
}

// DPI helper
func DPI(dpi float64) func(float64) int {
	return func(v float64) int {
		return int(v * dpi / 72.0)
	}
}

// Pixels helper
func Px(v int) int { return v }

// A4Width helper
func A4Width(dpi float64) SizeFunction {
	pixels := int(210.0 / 25.4 * dpi)
	return Fixed(pixels)
}

// A4Height helper
func A4Height(dpi float64) SizeFunction {
	pixels := int(297.0 / 25.4 * dpi)
	return Fixed(pixels)
}

// LayoutResult holds the result of the text layout
type LayoutResult struct {
	Lines          []Line
	PageSize       image.Point
	ContentStart   image.Point
	Margin         SpecMargin
	PageBackground color.Color
}

// TextToSpecs performs layout based on complex constraints.
// It returns the layout result containing lines, page size, and offsets.
func (sw *SimpleWrapper) TextToSpecs(opts ...SpecOption) (*LayoutResult, error) {
	config := SpecConfig{
		WidthFn:  Unbounded(),
		HeightFn: Unbounded(),
	}
	for _, opt := range opts {
		opt.ApplySpec(&config)
	}

	sw.boxer.Reset()
	// ... (layout logic, unchanged)
	inf := 1000000
	lines, _, err := sw.TextToRect(image.Rect(0, 0, inf, inf))
	if err != nil {
		return nil, fmt.Errorf("measure pass failed: %w", err)
	}
	naturalContentWidth := 0
	for _, l := range lines {
		s := l.Size()
		if s.Dx() > naturalContentWidth {
			naturalContentWidth = s.Dx()
		}
	}

	marginH := config.Margin.Left + config.Margin.Right
	marginV := config.Margin.Top + config.Margin.Bottom

	// Pass 2: Layout with width constraints and unconstrained height
	targetPageWidth := config.WidthFn(naturalContentWidth + marginH)
	if targetPageWidth < marginH+1 {
		targetPageWidth = marginH + 1
	}
	targetContentWidth := targetPageWidth - marginH

	// We use a large height for layout to detect natural height after wrapping
	// Then we apply HeightFn to constraint the final PageSize
	layoutHeight := 1000000

	sw.boxer.Reset()
	lines2, p, err := sw.TextToRect(image.Rect(0, 0, targetContentWidth, layoutHeight))
	if err != nil {
		return nil, fmt.Errorf("layout pass failed: %w", err)
	}
	naturalContentHeight := p.Y

	targetPageHeight := config.HeightFn(naturalContentHeight + marginV)
	if targetPageHeight < marginV+1 {
		targetPageHeight = marginV + 1
	}
	targetContentHeight := targetPageHeight - marginV

	if targetContentHeight < naturalContentHeight {
		sw.boxer.Reset()
		lines2, _, err = sw.TextToRect(image.Rect(0, 0, targetContentWidth, targetContentHeight))
		if err != nil {
			return nil, fmt.Errorf("final layout pass failed: %w", err)
		}
	}

	return &LayoutResult{
		Lines:          lines2,
		PageSize:       image.Point{X: targetPageWidth, Y: targetPageHeight},
		ContentStart:   image.Point{X: config.Margin.Left, Y: config.Margin.Top},
		Margin:         config.Margin,
		PageBackground: config.PageBackground,
	}, nil
}
