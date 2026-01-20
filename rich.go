package wordwrap

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// BackgroundPositioning defines how the background is positioned
type BackgroundPositioning int

const (
	BgPositioningSection5Zeroed BackgroundPositioning = iota // Aligned to content (inner) box 0,0
	BgPositioningZeroed                                      // Aligned to box 0,0 (Frame)
	BgPositioningPassThrough                                 // Absolute Coordinates (matches dst)
)

// FontColor defines the text color
type FontColor struct {
	color.Color
}

// BackgroundImage defines the background
type BackgroundImage struct {
	image.Image
	Positioning *BackgroundPositioning
	Fixed       *bool
}

// ... (existing code)

// BgImage returns a Group with BackgroundImage applied, or the Option if no args
func BgImage(i image.Image, args ...interface{}) interface{} {
	if len(args) == 0 {
		return BackgroundImage{Image: i}
	}
	var post []interface{}

	bi := BackgroundImage{Image: i}

	for _, a := range args {
		switch v := a.(type) {
		case BackgroundPositioning:
			bp := v
			bi.Positioning = &bp
		case BackgroundPositioningOption:
			bp := BackgroundPositioning(v)
			bi.Positioning = &bp
		case FixedBackgroundOption:
			b := bool(v)
			bi.Fixed = &b
		default:
			post = append(post, a)
		}
	}

	if len(post) > 0 {
		// Scoped usage (container)
		return Group{Args: append([]interface{}{bi}, post...)}
	}

	// Modifier usage (unscoped)
	return bi
}

// ... (existing code, helper functions)

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

// ContainerGroup defines a group that becomes a container box
type ContainerGroup struct {
	Options []interface{}
	Args    []interface{}
}

// Align defines an alignment
type BaselineAlignmentOption BaselineAlignment

// Helper functions (optional but requested)

// TextColor returns a Group with FontColor applied, or the Option if no args
func TextColor(c color.Color, args ...interface{}) interface{} {
	if len(args) == 0 {
		return FontColor{c}
	}
	return Group{Args: append([]interface{}{FontColor{c}}, args...)}
}

// TextImage returns a Group with FontImage applied, or the Option if no args
func TextImage(i image.Image, args ...interface{}) interface{} {
	if len(args) == 0 {
		return FontImage{i}
	}
	return Group{Args: append([]interface{}{FontImage{i}}, args...)}
}

// BgColor returns a Group with BackgroundColor applied, or the Option if no args
func BgColor(c color.Color, args ...interface{}) interface{} {
	if len(args) == 0 {
		return BackgroundImage{Image: image.NewUniform(c)}
	}
	return Group{Args: append([]interface{}{BackgroundImage{Image: image.NewUniform(c)}}, args...)}
}

type BackgroundPositioningOption BackgroundPositioning

// BgPosition returns a BackgroundPositioningOption
func BgPosition(p BackgroundPositioning) interface{} {
	return BackgroundPositioningOption(p)
}

// FixedBackground returns a Group with FixedBackground applied, or the Option if no args
func FixedBackground(args ...interface{}) interface{} {
	if len(args) == 0 {
		return FixedBackgroundOption(true)
	}
	var pre []interface{}
	var post []interface{}
	for _, a := range args {
		switch a.(type) {
		case BackgroundPositioning, BackgroundPositioningOption, FixedBackgroundOption:
			pre = append(pre, a)
		default:
			post = append(post, a)
		}
	}
	return Group{Args: append(append(pre, FixedBackgroundOption(true)), post...)}
}

// Container returns a ContainerGroup
func Container(args ...interface{}) interface{} {
	return ContainerGroup{Args: args}
}

type MarginOption fixed.Rectangle26_6
type PaddingOption fixed.Rectangle26_6
type IDOption struct{ ID interface{} }
type FixedBackgroundOption bool
type MinSizeOption fixed.Point26_6
type ResetOption struct{}

// Reset returns a ResetOption to plain style
func Reset() interface{} {
	return ResetOption{}
}

// Alignment returns a BaselineAlignmentOption
func Alignment(a BaselineAlignment) BaselineAlignmentOption { return BaselineAlignmentOption(a) }

// Highlight returns a Group with BackgroundColor applied (Alias for BgColor)
func Highlight(c color.Color, args ...interface{}) interface{} {
	return BgColor(c, args...)
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
			base := r.Min.Y + m.Ascent.Ceil() + 2 // +2 for offset?
			lineR := image.Rect(r.Min.X, base, r.Max.X, base+1)
			draw.Draw(i, lineR, &image.Uniform{c}, image.Point{}, draw.Over)
		},
	}
}

// MinWidth returns a MinSizeOption
func MinWidth(w int) interface{} {
	return MinSizeOption{X: fixed.I(w)}
}

// Align returns a Group with BaselineAlignmentOption applied
func Align(a BaselineAlignment, args ...interface{}) interface{} {
	return Group{Args: append([]interface{}{BaselineAlignmentOption(a)}, args...)}
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

	currentFont           font.Face
	defaultFont           font.Face
	currentStyle          *Style
	currentID             interface{}
	inBorder              bool
	currentDecoratorTypes []string
}

func (s *rcState) cloneStyle() *Style {
	if s.currentStyle == nil {
		return &Style{}
	}
	cp := *s.currentStyle
	// Copy slices
	if len(s.currentStyle.Effects) > 0 {
		cp.Effects = make([]BoxEffect, len(s.currentStyle.Effects))
		copy(cp.Effects, s.currentStyle.Effects)
	}
	if len(s.currentStyle.Decorators) > 0 {
		cp.Decorators = make([]func(Box) Box, len(s.currentStyle.Decorators))
		copy(cp.Decorators, s.currentStyle.Decorators)
	}
	return &cp
}

// BorderGroup implies args apply to Border
type BorderGroup Group
type BorderOption fixed.Rectangle26_6

// Border returns a BorderGroup with Border applied
func Border(rect fixed.Rectangle26_6, args ...interface{}) interface{} {
	return BorderGroup{Args: append([]interface{}{BorderOption(rect)}, args...)}
}

// Margin returns a Group with Margin applied, or the Option if no args
func Margin(rect fixed.Rectangle26_6, args ...interface{}) interface{} {
	if len(args) == 0 {
		return MarginOption(rect)
	}
	return Group{Args: append([]interface{}{MarginOption(rect)}, args...)}
}

// BoxPadding returns a Group with Padding applied, or the Option if no args
func BoxPadding(rect fixed.Rectangle26_6, args ...interface{}) interface{} {
	if len(args) == 0 {
		return PaddingOption(rect)
	}
	return Group{Args: append([]interface{}{PaddingOption(rect)}, args...)}
}

// ID returns a Group with ID applied, or the Option if no args
func ID(id interface{}, args ...interface{}) interface{} {
	if len(args) == 0 {
		return IDOption{ID: id}
	}
	return Group{Args: append([]interface{}{IDOption{ID: id}}, args...)}
}

// BackgroundColor returns a Group with BackgroundColor applied, or the Option if no args
func BackgroundColor(c color.Color, args ...interface{}) interface{} {
	return BgColor(c, args...)
}

// Color returns a Group with FontColor applied, or the Option if no args
func Color(c color.Color, args ...interface{}) interface{} {
	return TextColor(c, args...)
}

func (s *rcState) process(args []interface{}) {
	for _, arg := range args {
		switch v := arg.(type) {
		case Group:
			prevStyle := s.currentStyle
			prevFont := s.currentFont
			prevID := s.currentID
			prevInBorder := s.inBorder
			prevDecoratorTypes := s.currentDecoratorTypes

			s.currentStyle = s.cloneStyle()
			s.currentDecoratorTypes = append([]string(nil), s.currentDecoratorTypes...)
			s.process(v.Args)

			s.currentStyle = prevStyle
			s.currentFont = prevFont
			s.currentID = prevID
			s.inBorder = prevInBorder
			s.currentDecoratorTypes = prevDecoratorTypes

		case ContainerGroup:
			// logic: isolate contents
			// Children should NOT inherit decorators (Margin/Padding/Bg) from the container's scope.
			// But they SHOULD inherit Font/Color.

			// Clone state for children
			subS := &rcState{
				currentStyle:          s.cloneStyle(),
				currentFont:           s.currentFont,
				defaultFont:           s.defaultFont,
				drawer:                s.drawer,
				boxerOptions:          s.boxerOptions,
				tokenizer:             s.tokenizer,
				currentDecoratorTypes: append([]string(nil), s.currentDecoratorTypes...),
			}

			// Clear decorators for children so they don't get double-wrapped
			if subS.currentStyle != nil {
				subS.currentStyle.Decorators = nil
				subS.currentDecoratorTypes = nil
				// Keep Effects? Probably yes.
				// Keep Alignment? Maybe.
			}

			// Process children
			subS.process(v.Args)

			// Apply PARENT decorators to the Container itself
			var opts []ContentOption
			if s.currentStyle != nil {
				if len(s.currentStyle.Decorators) > 0 {
					opts = append(opts, WithDecorators(s.currentStyle.Decorators...))
				}
				if len(s.currentStyle.Effects) > 0 {
					opts = append(opts, WithBoxEffects(s.currentStyle.Effects))
				}
				if s.currentStyle.Alignment != AlignBaseline {
					opts = append(opts, WithAlignment(s.currentStyle.Alignment))
				}
			}
			if s.currentID != nil {
				opts = append(opts, WithID(s.currentID))
			}

			c := NewContainerContent(subS.contents, opts...)
			s.contents = append(s.contents, c)

		case BorderGroup:
			prevStyle := s.currentStyle
			prevFont := s.currentFont
			prevID := s.currentID
			prevInBorder := s.inBorder
			prevDecoratorTypes := s.currentDecoratorTypes

			s.currentStyle = s.cloneStyle()
			s.currentDecoratorTypes = append([]string(nil), s.currentDecoratorTypes...)
			s.inBorder = true
			s.process(v.Args)

			s.currentStyle = prevStyle
			s.currentFont = prevFont
			s.currentID = prevID
			s.inBorder = prevInBorder
			s.currentDecoratorTypes = prevDecoratorTypes
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
			// Use Decorator Logic
			bg := v.Image

			// Resolve Fixed
			fixedBg := s.currentStyle.FixedBackground
			if v.Fixed != nil {
				fixedBg = *v.Fixed
			}

			// Resolve Positioning
			bgPos := s.currentStyle.BgPositioning
			if v.Positioning != nil {
				bgPos = *v.Positioning
			}

			// Map legacy FixedBackground to Positioning if not set (or if explicitly requested via Fixed=true)
			// Logic: If legacy Fixed is true, and Pos is default (0), assume PassThrough.
			if fixedBg && bgPos == BgPositioningSection5Zeroed {
				bgPos = BgPositioningPassThrough
			}

			d := func(b Box) Box {
				// DecorationBox supports Background.
				return NewDecorationBox(b, fixed.Rectangle26_6{}, fixed.Rectangle26_6{}, bg, bgPos)
			}
			idx := -1
			for i, t := range s.currentDecoratorTypes {
				if t == "Background" {
					idx = i
					break
				}
			}
			if idx != -1 {
				s.currentStyle.Decorators[idx] = d
			} else {
				s.currentStyle.Decorators = append(s.currentStyle.Decorators, d)
				s.currentDecoratorTypes = append(s.currentDecoratorTypes, "Background")
			}

		case MarginOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			margin := fixed.Rectangle26_6(v)
			bgPos := s.currentStyle.BgPositioning
			if s.currentStyle.FixedBackground && bgPos == BgPositioningSection5Zeroed {
				bgPos = BgPositioningPassThrough
			}
			d := func(b Box) Box {
				// We pass nil bg here because BgImage logic handles background separately now?
				// User wants separate layers.
				return NewDecorationBox(b, fixed.Rectangle26_6{}, margin, nil, bgPos) // fixed irrelevant if bg nil
			}
			s.currentStyle.Decorators = append(s.currentStyle.Decorators, d)
			s.currentDecoratorTypes = append(s.currentDecoratorTypes, "Margin")

		case PaddingOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			padding := fixed.Rectangle26_6(v)
			bgPos := s.currentStyle.BgPositioning
			if s.currentStyle.FixedBackground && bgPos == BgPositioningSection5Zeroed {
				bgPos = BgPositioningPassThrough
			}
			d := func(b Box) Box {
				return NewDecorationBox(b, padding, fixed.Rectangle26_6{}, nil, bgPos)
			}
			s.currentStyle.Decorators = append(s.currentStyle.Decorators, d)
			s.currentDecoratorTypes = append(s.currentDecoratorTypes, "Padding")

		case BorderOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			// Border should use BorderImage if set?
			// BorderOption defines rect. BorderGroup sets logic.
			// But user might use Border(rect, BgImage(borderImg)).
			// That would append BgDecorator then BorderDecorator.
			// We want BorderDecorator to use the image?
			// Currently BgImage adds BgDecorator.
			// If Border(rect) wraps arg, and arg is BgImage.
			// Arg adds Bg Decorator.
			// Then BorderOption adds Border Decorator.
			// SimpleBoxer Reverse: Border(Bg(Box)).
			// Bg(Box) is a box with BorderImage (via standard Bg logic).
			// Border(Box) adds border space (padding/margin?).
			// If Border means "Frame", we usually want the image ON the border.
			// NewDecorationBox supports this if we pass Margin=BorderWidth and Bg=BorderImage.
			// But here we split them.
			// Bg(Box) -> DecorationBox(bg=BorderImage).
			// Border(Box) -> DecorationBox(Margin=BorderWidth).
			// Result: DecorationBox(Margin) ( DecorationBox(Bg) ( Box ) ).
			// Visual: Outer Margin (Transp). Then Inner Box (Bg).
			// But we want Bg to be IN the Margin?
			// `DecorationBox` draws Bg covering Margin+Padding.
			// If we split, `MarginDecorator` draws nothing. `BgDecorator` draws inside.
			// So `Bg` is INSIDE `Margin`.
			// This matches "Border" concept?
			// No, `Margin` is external spacing.
			// `Border` is usually visualized frame.
			// If user wants `Border(5px, Bg(Wood))`.
			// `Bg(Wood)` creates WoodBox.
			// `Border(5px)` creates SpaceBox wrapping WoodBox.
			// Result: `[ Space [ Wood [ Text ] ] ]`.
			// Visual: Transparent Border around Wood Box.
			// THIS IS WRONG. User wants Wood Frame.
			// So `BgImage` MUST apply to the `BorderDecorator`?
			// OR we must COMBINE them.
			// Since `BgImage` helper appends Decorator.
			// We can't easily combine later without inspecting list.

			// Alternative: `BgImage` checks if `inBorder`?
			// If `inBorder`, it sets `s.currentStyle.BorderImage` state instead of adding decorator.
			// Then `BorderOption` checks `BorderImage` state and creates combined decorator.
			// `Border(rect, Bg(img))`.
			// `Bg` runs (inside BorderGroup). Sets `s.currentStyle.BorderImage = img`.
			// `BorderOption` runs (last). Checks `BorderImage`. Creates `DecorationBox(Margin=rect, Bg=BorderImage)`.
			// This works!
			// So `Border` helper must use `BorderGroup` logic correctly.
			// `Border` helper returns `BorderGroup`.
			// `process(BorderGroup)` sets `inBorder=true`.
			// `BgImage` helper MUST NOT add decorator if `inBorder`.

			margin := fixed.Rectangle26_6(v)
			bg := s.currentStyle.BorderImage
			bgPos := s.currentStyle.BgPositioning
			if s.currentStyle.FixedBackground && bgPos == BgPositioningSection5Zeroed {
				bgPos = BgPositioningPassThrough
			}
			d := func(b Box) Box {
				return NewDecorationBox(b, fixed.Rectangle26_6{}, margin, bg, bgPos)
			}
			s.currentStyle.Decorators = append(s.currentStyle.Decorators, d)
			s.currentDecoratorTypes = append(s.currentDecoratorTypes, "Border")

		case BackgroundPositioningOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.BgPositioning = BackgroundPositioning(v)
			// Sync legacy
			if BackgroundPositioning(v) == BgPositioningPassThrough {
				s.currentStyle.FixedBackground = true
			} else {
				s.currentStyle.FixedBackground = false
			}
		case BackgroundPositioning:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.BgPositioning = v
			// Sync legacy
			if v == BgPositioningPassThrough {
				s.currentStyle.FixedBackground = true
			} else {
				s.currentStyle.FixedBackground = false
			}
		case FixedBackgroundOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			s.currentStyle.FixedBackground = bool(v)
		case MinSizeOption:
			if s.currentStyle == nil {
				s.currentStyle = &Style{}
			}
			minSize := fixed.Point26_6(v)
			s.currentStyle.MinSize = minSize
			d := func(b Box) Box {
				return &MinSizeBox{Box: b, MinSizeVal: minSize}
			}
			s.currentStyle.Decorators = append(s.currentStyle.Decorators, d)
			s.currentDecoratorTypes = append(s.currentDecoratorTypes, "MinSize")
		case ResetOption:
			s.currentStyle = &Style{}
			s.currentDecoratorTypes = nil
			s.currentFont = s.defaultFont
			s.currentID = nil
			s.inBorder = false
			if s.defaultFont != nil {
				s.currentStyle.font = s.defaultFont
			}
		case IDOption:
			s.currentID = v.ID
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
				if s.currentStyle.Alignment != AlignBaseline {
					opts = append(opts, WithAlignment(s.currentStyle.Alignment))
				}
				if len(s.currentStyle.Decorators) > 0 {
					opts = append(opts, WithDecorators(s.currentStyle.Decorators...))
				}
				if s.currentStyle.MinSize != (fixed.Point26_6{}) {
					opts = append(opts, WithMinSize(s.currentStyle.MinSize))
				}
			}
			if len(s.currentStyle.Effects) > 0 {
				opts = append(opts, WithBoxEffects(s.currentStyle.Effects))
			}
			if s.currentID != nil {
				opts = append(opts, WithID(s.currentID))
			}

			c := NewContent(v, opts...)
			s.contents = append(s.contents, c)
		case ImageContent:
			var opts []ContentOption
			if s.currentStyle != nil {
				if s.currentStyle.Alignment != AlignBaseline {
					opts = append(opts, WithAlignment(s.currentStyle.Alignment))
				}
				if len(s.currentStyle.Decorators) > 0 {
					opts = append(opts, WithDecorators(s.currentStyle.Decorators...))
				}
			}
			if len(s.currentStyle.Effects) > 0 {
				opts = append(opts, WithBoxEffects(s.currentStyle.Effects))
			}
			if s.currentID != nil {
				opts = append(opts, WithID(s.currentID))
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
