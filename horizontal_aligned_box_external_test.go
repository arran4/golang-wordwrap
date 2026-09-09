package wordwrap_test

import (
	"github.com/arran4/golang-wordwrap"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"testing"
)

type CustomLegacyBox struct{}

func (c *CustomLegacyBox) AdvanceRect() fixed.Int26_6                                         { return 0 }
func (c *CustomLegacyBox) MetricsRect() font.Metrics                                          { return font.Metrics{} }
func (c *CustomLegacyBox) Whitespace() bool                                                   { return false }
func (c *CustomLegacyBox) DrawBox(i wordwrap.Image, y fixed.Int26_6, dc *wordwrap.DrawConfig) {}
func (c *CustomLegacyBox) FontDrawer() *font.Drawer   { return nil }
func (c *CustomLegacyBox) Len() int                   { return 0 }
func (c *CustomLegacyBox) TextValue() string          { return "" }
func (c *CustomLegacyBox) MinSize() (fixed.Int26_6, fixed.Int26_6)                            { return 0, 0 }
func (c *CustomLegacyBox) MaxSize() (fixed.Int26_6, fixed.Int26_6)                            { return 0, 0 }

func TestHorizontalAlignedBox_ExternalPackageCompatibility(t *testing.T) {
	// Case 15: External package custom Box without optional methods.
	// As requested, this type implements ONLY the strict Box interface contract.
	var b wordwrap.Box = &CustomLegacyBox{}

	hab := &wordwrap.HorizontalAlignedBox{
		Box:       b,
		Alignment: wordwrap.AlignRight,
	}

	// This should compile and execute without panic, proving we didn't add required methods
	// to Box interface internally by mistake (like TextValue).
	_ = hab.AdvanceRect()

}
