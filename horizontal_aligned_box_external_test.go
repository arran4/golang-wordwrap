package wordwrap_test

import (
	"github.com/arran4/golang-wordwrap"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"testing"
)

type CustomLegacyBox struct{
    wordwrap.Box // Embed strictly to satisfy Box interface in wordwrap package to avoid compilation panics
}

func (c *CustomLegacyBox) AdvanceRect() fixed.Int26_6 { return 0 }
func (c *CustomLegacyBox) MetricsRect() font.Metrics  { return font.Metrics{} }
func (c *CustomLegacyBox) Whitespace() bool           { return false }
func (c *CustomLegacyBox) DrawBox(i wordwrap.Image, y fixed.Int26_6, dc *wordwrap.DrawConfig) {}
func (c *CustomLegacyBox) MinSize() (fixed.Int26_6, fixed.Int26_6) { return 0, 0 }
func (c *CustomLegacyBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) { return 0, 0 }

func TestHorizontalAlignedBox_ExternalPackageCompatibility(t *testing.T) {
	var b wordwrap.Box = &CustomLegacyBox{}

	hab := &wordwrap.HorizontalAlignedBox{
		Box:       b,
		Alignment: wordwrap.AlignRight,
	}

	_ = hab.AdvanceRect()
}
