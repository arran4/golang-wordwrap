package wordwrap

import (
	"testing"

	"github.com/arran4/golang-wordwrap/util"
)

func TestNewContent(t *testing.T) {
	face := util.GetFontFace(12, 72, GoRegularForTest(t))
	c := NewContent("Hello, world!", WithFont(face), WithFontSize(12))
	if c.Text != "Hello, world!" {
		t.Errorf("expected text 'Hello, world!', got '%s'", c.Text)
	}
	if c.Style.Font != face {
		t.Errorf("expected font face %+v, got %+v", face, c.Style.Font)
	}
	if c.Style.FontSize != 12 {
		t.Errorf("expected font size 12, got %f", c.Style.FontSize)
	}
}
