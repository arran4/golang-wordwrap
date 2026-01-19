package wordwrap

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// BaselineAlignment alignment of content within the line
type BaselineAlignment int

const (
	AlignBaseline BaselineAlignment = iota
	AlignTop
	AlignMiddle
	AlignBottom
)

// Content is a struct that holds the text and style of a piece of content.
type Content struct {
	text       string
	style      *Style
	image      image.Image
	id         interface{}
	imageScale float64
	decorators []func(Box) Box
	children   []*Content
}

// Style is a struct that holds the font and font size of a piece of content.
type Style struct {
	font            font.Face
	FontDrawerSrc   image.Image
	BackgroundColor image.Image // can be used for colour or image
	Padding         fixed.Rectangle26_6
	Margin          fixed.Rectangle26_6
	Alignment       BaselineAlignment
	Effects         []BoxEffect
	FixedBackground bool
	Border          fixed.Rectangle26_6
	BorderImage     image.Image
	Decorators      []func(Box) Box
	MinSize         fixed.Point26_6
}

// WithMinSize sets the minimum size of the content
func WithMinSize(size fixed.Point26_6) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = NewStyle()
		}
		c.style.MinSize = size
	}
}

// NewStyle creates a new style
func NewStyle() *Style {
	return &Style{}
}

// NewContent creates a new Content object with the given text and options.
func NewContent(text string, opts ...ContentOption) *Content {
	c := &Content{
		text: text,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewImageContent creates a new Content object with the given image and options.
func NewImageContent(i image.Image, opts ...ContentOption) *Content {
	c := &Content{
		image: i,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ContentOption is a function that can be used to configure a Content object.
type ContentOption func(*Content)

// WithFont sets the font of a Content object.
func WithFont(font font.Face) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.font = font
	}
}

// WithFontColor sets the font color of a Content object.
func WithFontColor(col color.Color) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.FontDrawerSrc = image.NewUniform(col)
	}
}

// WithFontImage sets the font image of a Content object.
func WithFontImage(i image.Image) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.FontDrawerSrc = i
	}
}

// WithAlignment sets the vertical alignment of a Content object.
func WithAlignment(a BaselineAlignment) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.Alignment = a
	}
}

// WithBackgroundColor sets the background color of a Content object.
func WithBackgroundColor(col color.Color) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.BackgroundColor = image.NewUniform(col)
	}
}

// WithPadding sets the padding of a Content object.
func WithPadding(p fixed.Rectangle26_6) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = NewStyle()
		}
		c.style.Padding = p
	}
}

// WithMargin sets the margin of a Content object.
func WithMargin(m fixed.Rectangle26_6) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = NewStyle()
		}
		c.style.Margin = m
	}
}

// WithBackendImage sets the background image of a Content object.
func WithBackendImage(i image.Image) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = NewStyle()
		}
		c.style.BackgroundColor = i
	}
}

// WithID sets the ID of a Content object.
func WithID(id interface{}) ContentOption {
	return func(c *Content) {
		c.id = id
	}
}

// ID returns the ID of the content
func (c *Content) ID() interface{} {
	return c.id
}

// WithFixedBackground sets whether the background is "fixed" (global coordinates)
func WithFixedBackground(fixed bool) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = NewStyle()
		}
		c.style.FixedBackground = fixed
	}
}

// WithImageScale sets the scale of an image Content object.
func WithImageScale(s float64) ContentOption {
	return func(c *Content) {
		c.imageScale = s
	}
}

// WithDecorators adds decorators to the content
func WithDecorators(ds ...func(Box) Box) ContentOption {
	return func(c *Content) {
		c.decorators = append(c.decorators, ds...)
	}
}

// NewContainerContent creates a new container Content object.
func NewContainerContent(children []*Content, opts ...ContentOption) *Content {
	c := &Content{
		children: children,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
