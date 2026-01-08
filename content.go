package wordwrap

import (
	"golang.org/x/image/font"
)

// Content represents a piece of styled text.
type Content struct {
	Text  string
	Style Style
}

// Style represents the style of a piece of text.
type Style struct {
	Font     font.Face
	FontSize float64
}

// NewContent creates a new Content object.
func NewContent(text string, options ...ContentOption) *Content {
	c := &Content{
		Text: text,
	}
	for _, option := range options {
		option.ApplyContentOption(c)
	}
	return c
}

// ContentOption is an option for a Content object.
type ContentOption interface {
	ApplyContentOption(*Content)
}
