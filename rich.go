package wordwrap

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
)

// FontColor defines the text color
type FontColor struct {
	color.Color
}

// BackgroundImage defines the background
type BackgroundImage struct {
	image.Image
}

// ImageContent defines an inline image
type ImageContent struct {
	Image image.Image
	Scale float64 // 0 or 1 = original size
}

// Helper functions (optional but requested)

// FontImage defines the text pattern
type FontImage struct {
	image.Image
}

// Group defines a group of arguments
type Group struct {
	Args []interface{}
}

// Align defines an alignment
type BaselineAlignmentOption BaselineAlignment

// Helper functions (optional but requested)

// TextColor returns a FontColor
func TextColor(c color.Color) FontColor { return FontColor{c} }

// TextImage returns a FontImage
func TextImage(i image.Image) FontImage { return FontImage{i} }

// BgColor returns a BackgroundImage of a uniform color
func BgColor(c color.Color) BackgroundImage { return BackgroundImage{image.NewUniform(c)} }

// BgImage returns a BackgroundImage
func BgImage(i image.Image) BackgroundImage { return BackgroundImage{i} }

// Alignment returns a BaselineAlignmentOption
func Alignment(a BaselineAlignment) BaselineAlignmentOption { return BaselineAlignmentOption(a) }

// Highlight returns a BoxEffect for highlighting
func Highlight(c color.Color) BoxEffect {
	return BoxEffect{
		Type: EffectPre,
		Func: func(i Image, b Box, dc *DrawConfig) {
			r := i.Bounds()
			draw.Draw(i, r, &image.Uniform{c}, image.Point{}, draw.Over)
		},
	}
}

// Strikethrough returns a Post effect
func Strikethrough(c color.Color) BoxEffect {
	return BoxEffect{
		Type: EffectPost,
		Func: func(i Image, b Box, dc *DrawConfig) {
			r := i.Bounds()
			m := b.MetricsRect()

			mid := r.Min.Y + m.Ascent.Ceil()/2 + (m.Ascent.Ceil() / 4) // Roughly middle of X-height
			lineR := image.Rect(r.Min.X, mid, r.Max.X, mid+1)
			draw.Draw(i, lineR, &image.Uniform{c}, image.Point{}, draw.Over)
		},
	}
}

// Underline returns a Post effect
func Underline(c color.Color) BoxEffect {
	return BoxEffect{
		Type: EffectPost,
		Func: func(i Image, b Box, dc *DrawConfig) {
			r := i.Bounds()
			m := b.MetricsRect()
			// Baseline is approximately m.Ascent relative to top?
			// DrawBox passed 'y' which is line baseline.
			// But here we are drawing on 'i' (subimage).
			// If box is aligned to baseline, 'y' in DrawBox was likely used to offset dot.
			// But 'i' bounds are the box's allocated rect.
			// Text is drawn relative to 'i.Min.Y + y'.
			// But we don't have 'y'.
			// However simple text box draws at 'b.Min.Y + y'.
			// Does simple text box shift 'i'? No.
			// The content is drawn.
			// The logical baseline of the text inside the box is at m.Ascent from the top of the box.
			base := r.Min.Y + m.Ascent.Ceil() + 2 // +2 for offset?
			lineR := image.Rect(r.Min.X, base, r.Max.X, base+1)
			draw.Draw(i, lineR, &image.Uniform{c}, image.Point{}, draw.Over)
		},
	}
}

// ProcessRichArgs parses variadic arguments into standard components
func ProcessRichArgs(args ...interface{}) ([]*Content, *font.Drawer, []WrapperOption, []BoxerOption, Boxer, Tokenizer) {
	state := &rcState{
		currentStyle: &Style{},
	}
	state.process(args)
	if state.drawer == nil && state.defaultFont != nil {
		state.drawer = &font.Drawer{
			Src:  image.NewUniform(image.Black),
			Face: state.defaultFont,
		}
	}
	return state.contents, state.drawer, state.wrapperOptions, state.boxerOptions, state.boxer, state.tokenizer
}

type rcState struct {
	contents       []*Content
	drawer         *font.Drawer
	wrapperOptions []WrapperOption
	boxerOptions   []BoxerOption
	boxer          Boxer
	tokenizer      Tokenizer

	currentFont  font.Face
	defaultFont  font.Face
	currentStyle *Style
}

func (s *rcState) cloneStyle() *Style {
	if s.currentStyle == nil {
		return &Style{}
	}
	cp := *s.currentStyle
	// Copy slice
	if len(s.currentStyle.Effects) > 0 {
		cp.Effects = make([]BoxEffect, len(s.currentStyle.Effects))
		copy(cp.Effects, s.currentStyle.Effects)
	}
	return &cp
}

func (s *rcState) process(args []interface{}) {
	for _, arg := range args {
		switch v := arg.(type) {
		case Group:
			// Save state
			prevStyle := s.currentStyle
			prevFont := s.currentFont
			// New scope (clone style)
			s.currentStyle = s.cloneStyle()
			// Process children
			s.process(v.Args)
			// Restore state
			s.currentStyle = prevStyle
			s.currentFont = prevFont
		case []*Content:
			s.contents = append(s.contents, v...)
		case *Content:
			s.contents = append(s.contents, v)
		case font.Face:
			s.currentFont = v
			s.updateCurrentStyle()
			if s.defaultFont == nil {
				s.defaultFont = v
			}
		case FontColor:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.FontDrawerSrc = image.NewUniform(v.Color)
		case FontImage:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.FontDrawerSrc = v.Image
		case BackgroundImage:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.BackgroundColor = v.Image
		case BoxEffect:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.Effects = append(s.currentStyle.Effects, v)
		case BaselineAlignmentOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.Alignment = BaselineAlignment(v)
		case *font.Drawer:
			if v != nil {
				s.drawer = v
				s.currentFont = v.Face
				s.updateCurrentStyle()
				if s.defaultFont == nil {
					s.defaultFont = v.Face
				}
			}
		case string:
			var opts []ContentOption
			if s.currentStyle != nil {
				if s.currentStyle.font != nil {
					opts = append(opts, WithFont(s.currentStyle.font))
				}
				if s.currentStyle.FontDrawerSrc != nil {
					opts = append(opts, WithFontImage(s.currentStyle.FontDrawerSrc))
				}
				if s.currentStyle.BackgroundColor != nil {
					opts = append(opts, WithBackendImage(s.currentStyle.BackgroundColor))
				}
				if s.currentStyle.Alignment != AlignBaseline {
					opts = append(opts, WithAlignment(s.currentStyle.Alignment))
				}
			}
			if len(s.currentStyle.Effects) > 0 {
				opts = append(opts, WithBoxEffects(s.currentStyle.Effects))
			}

			c := NewContent(v, opts...)
			s.contents = append(s.contents, c)
		case ImageContent:
			var opts []ContentOption
			if s.currentStyle != nil {
				if s.currentStyle.BackgroundColor != nil {
					opts = append(opts, WithBackendImage(s.currentStyle.BackgroundColor))
				}
				if s.currentStyle.Alignment != AlignBaseline {
					opts = append(opts, WithAlignment(s.currentStyle.Alignment))
				}
				if len(s.currentStyle.Effects) > 0 {
					opts = append(opts, WithBoxEffects(s.currentStyle.Effects))
				}
			}
			if v.Scale != 0 {
				opts = append(opts, WithImageScale(v.Scale))
			}
			c := NewImageContent(v.Image, opts...)
			s.contents = append(s.contents, c)
		case BoxerOption:
			s.boxerOptions = append(s.boxerOptions, v)
		case WrapperOption:
			s.wrapperOptions = append(s.wrapperOptions, v)
		case Boxer:
			s.boxer = v
		case Tokenizer:
			s.tokenizer = v
		}
	}
}

// WithBoxEffects sets the effects
func WithBoxEffects(e []BoxEffect) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.Effects = append(c.style.Effects, e...)
	}
}

func (s *rcState) updateCurrentStyle() {
	if s.currentStyle == nil {
		s.currentStyle = &Style{}
	}
	if s.currentFont != nil {
		s.currentStyle.font = s.currentFont
	}
}
