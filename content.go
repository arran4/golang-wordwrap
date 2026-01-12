package wordwrap

import "golang.org/x/image/font"

// Content is a struct that holds the text and style of a piece of content.
type Content struct {
	text  string
	style *Style
}

// Style is a struct that holds the font and font size of a piece of content.
type Style struct {
	font font.Face
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
