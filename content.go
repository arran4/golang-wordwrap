package wordwrap

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
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
	imageScale float64
}

// Style is a struct that holds the font and font size of a piece of content.
type Style struct {
	font            font.Face
	FontDrawerSrc   image.Image
	BackgroundColor image.Image // can be used for colour or image
	Alignment       BaselineAlignment
	Effects         []BoxEffect
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

// WithBackendImage sets the background image of a Content object.
func WithBackendImage(i image.Image) ContentOption {
	return func(c *Content) {
		if c.style == nil {
			c.style = &Style{}
		}
		c.style.BackgroundColor = i
	}
}

// WithImageScale sets the scale of an image Content object.
func WithImageScale(s float64) ContentOption {
	return func(c *Content) {
		c.imageScale = s
	}
}
