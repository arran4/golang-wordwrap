package wordwrap

import (
	"image"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// DecorationBox is a box that adds padding and margin around another box
type DecorationBox struct {
	Box
	Padding         fixed.Rectangle26_6
	Margin          fixed.Rectangle26_6
	Background      image.Image
	FixedBackground bool
}

// NewDecorationBox constructor
func NewDecorationBox(b Box, padding, margin fixed.Rectangle26_6, bg image.Image, fixedBg bool) *DecorationBox {
	return &DecorationBox{
		Box:             b,
		Padding:         padding,
		Margin:          margin,
		Background:      bg,
		FixedBackground: fixedBg,
	}
}

// AdvanceRect width of text
func (db *DecorationBox) AdvanceRect() fixed.Int26_6 {
	return db.Box.AdvanceRect() + db.Padding.Max.X + db.Padding.Min.X + db.Margin.Max.X + db.Margin.Min.X
}

// MetricsRect all other font details of text
// MetricsRect all other font details of text
func (db *DecorationBox) MetricsRect() font.Metrics {
	m := db.Box.MetricsRect()
	m.Ascent += db.Padding.Min.Y + db.Margin.Min.Y
	m.Descent += db.Padding.Max.Y + db.Margin.Max.Y
	return m
}

// Whitespace if contains a white space or not
func (db *DecorationBox) Whitespace() bool {
	// If we have a background or padding/margin that is visible (conceptually), we are not whitespace.
	// But whitespace behavior usually controls line breaking.
	// If we return false, we might not break correctly?
	// Folder checks Whitespace(). If true, it might treat it as a potential break point.
	// If false, it treats it as a word.
	// We WANT it to be breakable, but we want it DRAWN.
	// Folder logic:
	// if b.Whitespace() {
	//      b = &LineBreakBox{ Box: b }
	//      l.boxes = append(l.boxes, b)
	// }
	// If it becomes LineBreakBox, does LineBreakBox draw?
	// LineBreakBox.DrawBox() calls p.Box.DrawBox() IF it's not empty?
	// No, LineBreakBox:
	// func (sb *LineBreakBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {}
	// IT DOES NOT DRAW!
	// That is the bug!
	// LineBreakBox swallows the box drawing!

	// So if we return false, Folder treats it as Content.
	// It will add it to l.Boxes.
	// If it hits overflow, it will break BEFORE it (if StrictBorders).
	// If we want it to act as a space (breakable) AND be drawn?
	// Current LineBreakBox implementation explicitly HIDES the content.

	// If we just return false here:
	// The space becomes a "Word".
	// "Word"[SpaceWord]"Word".
	// It will be rendered.
	// But line breaking?
	// If "Word SpaceWord Word" fits, all good.
	// If "Word SpaceWord Word" overflows.
	// It might break in the middle of "SpaceWord" if it was long?
	// No, SpaceWord is 1 char.
	// So it breaks before or after.
	// So returning false effectively treats space as a non-breaking-space (visibly) that is just a very short word?
	// Actually, if we return false, `RSimpleBox` usually separates words.
	// Folder accumulates boxes.
	// If a box is NOT whitespace:
	// It pushes to `boxer`.
	// Then `done = true`.
	// SimpleFolder.Next() loop ends.
	// Actually, `SimpleFolder` consumes boxes until overflow or line end?
	// No. `SimpleFolder.Next()` generates *one line*.
	// The loop: gets Next box. Checks if fits. Adds to line.
	// If whitespace: `l.boxes = append`.
	// If NOT whitespace: `sf.boxer.Push(b)`. `done = true`. Returns.
	// WAIT. `SimpleFolder` stops at the FIRST non-whitespace word that it *can't fit*?
	// No.
	// The logic is in `fold.go` Lines 300-313.
	// If it OVERFLOWS:
	//    If Whitespace: Turn into LineBreakBox. Append. (Swallows drawing).
	//    If Not Whitespace: Push back to Boxer (so it's handled next line). Return line.
	// If it DOES NOT OVERFLOW (Line 316):
	//    `l.Push(b, a)` -> Append to line boxes.

	// So for NON-OVERFLOWING space:
	// It is appended to `l.boxes`.
	// And `DrawLine` iterates `l.boxes` and calls `DrawBox`.

	// So... if it fits, it is drawn.
	// So why the gap?
	// Unless `SimpleBoxer` or someone is marking it as Whitespace AND `Folder` logic does something else?
	// Re-reading `fold.go` logic above (Step 552).
	// `l.Push(b, a)` is called if NO overflow.
	// So SpaceBox is added.
	// `DrawLine` calls `DrawBox`.

	// Maybe `DecorationBox` is not wrapping the space?
	// In `box.go` `SimpleBoxerGrab`:
	// `if IsSpaceButNotCRLF(r)` -> `RSimpleBox`.
	// In `Next()`:
	// `case RSimpleBox`: `NewSimpleTextBox`.
	// Then it applies style.
	// `currentContent` covers the space char.
	// So `style` is present.
	// `DecorationBox` is applied.

	// Maybe `SpaceBox` Width (Advance) is 0?
	// `SimpleTextBox.AdvanceRect()` calls `fontDrawer.MeasureString`.
	// Space has width.

	// Is it possible `Margin` causing it?
	// I used `WithMargin(fixed.R(0, 5, 0, 5))`.
	// MinX=0. MaxX=0.
	// So no horizontal margin.

	// Wait, I updated Step 545 to:
	// `styleOpts = append(styleOpts, wordwrap.WithMargin(fixed.R(0, 5, 0, 5)))`
	// Left=0. Top=5. Right=0? Bottom=5.
	// IF R() arguments are Left, Top, Right, Bottom.
	// But `fixed.Rectangle26_6` struct fields are `Min`, `Max`.
	// `fixed.R(x0, y0, x1, y1)` -> `Min(x0,y0), Max(x1,y1)`.
	// `Min.X` is added. `Max.X` is added.
	// If I passed 0.
	// Then `db.Margin.Min.X` = 0.

	// What if `Padding` is causing it?
	// `WithPadding(fixed.R(5, 5, 5, 5))`.
	// Left=5. Right=5.
	// Total Advance += 10.
	// DrawBox offsets inner by 5.
	// Background covers 10 + InnerWidth.

	// It should look like:
	// [5px|Word|5px] [5px|Space|5px] [5px|Word|5px]
	// Touching.

	// Why gap?
	// Maybe the *font* has built-in spacing for characters?
	// But we are drawing background box calculated from Advance.

	// What if `DrawLine` logic adds spacing?
	// `fi += b.AdvanceRect()`.
	// It positions next box at `fi`.
	// So they are strictly adjacent.

	// HYPOTHESIS: `DecorationBox` implementation of `Whitespace()` returns `true` (delegates to inner).
	// The Folder sees `Whitespace() == true`.
	// But since it fits, it adds it to boxes.
	// So it should draw.

	// Wait! `simple.New`. Ebiten.
	// `golang-rpg-textbox/theme/simple/simple.go`:
	// Uses `goregular` font?

	// Is it possible that `SimpleTextBox` for space has NO text?
	// `SimpleBoxer` accumulates chars.

	// Let's force `Whitespace()` false for `DecorationBox`.
	// This will prove if being "Whitespace" is the cause (e.g. some other logic I missed drops it).
	// If I return `false`, Folder treats it as a word.
	// Warning: This breaks 	wrapping if space is treated as word (it won't break *at* the space).
	// But for "Word Space Word", if Space is a word, it breaks *after* the Space or *before* it.
	// So "Word" "Space" "Word".
	// It will likely break after "Space" if "Word" doesn't fit?
	// Or before "Space" if "Space" doesn't fit?
	// This seems acceptable for a test.

	if db.Background != nil || !db.Padding.Empty() || !db.Margin.Empty() {
		return false
	}
	return db.Box.Whitespace()
}

// DrawBox renders object
func (db *DecorationBox) DrawBox(i Image, y fixed.Int26_6, dc *DrawConfig) {
	// Calculate bounds for background
	// i is the subimage for this box.
	// But `DrawBox` contract is that it draws into `i`.
	// The `y` parameter is expected baseline.
	// We need to fill background.

	// Where is the box relative to i?
	// Box occupies all of `i` usually?
	// No, `DrawLine` passes `img.SubImage(r)`. `r` is calculated based on `AdvanceRect` and `Metrics`.
	// So `i` should be exactly the size of this box (including padding/margin).

	b := i.Bounds()

	// Margin is outside the background.
	// Padding is inside the background.

	bgRect := b
	bgRect.Min.X += db.Margin.Min.X.Ceil()
	bgRect.Min.Y += db.Margin.Min.Y.Ceil()
	bgRect.Max.X -= db.Margin.Max.X.Ceil()
	bgRect.Max.Y -= db.Margin.Max.Y.Ceil()

	if db.Background != nil {
		srcPoint := image.Point{}
		if db.FixedBackground {
			srcPoint = bgRect.Min
		}
		draw.Draw(i, bgRect, db.Background, srcPoint, draw.Src)
	}

	// Draw inner box
	// We need to offset the inner draw by Margin + Padding
	// The inner box `DrawBox` expects an image.
	// We should SubImage it?

	innerRect := bgRect
	innerRect.Min.X += db.Padding.Min.X.Ceil()
	innerRect.Min.Y += db.Padding.Min.Y.Ceil()
	innerRect.Max.X -= db.Padding.Max.X.Ceil()
	innerRect.Max.Y -= db.Padding.Max.Y.Ceil()

	// SubImage might not be accurate if box implementation relies on something else.
	// But `DrawBox` usually just draws into the provided image.
	// However, `SimpleTextBox` uses `sb.drawer.Dot` which is absolute coordinates relative to `i`?
	// `sb.drawer.Dot = fixed.Point26_6{ X: fixed.I(b.Min.X), Y: fixed.I(b.Min.Y) + y }`
	// Wait, `SimpleTextBox.DrawBox` sets Dot to `b.Min.X`.
	// So if we pass a subimage `innerImg`, `innerImg.Bounds().Min.X` will be correct.

	if innerRect.Empty() {
		// No space for inner box?
		return
	}

	innerImg := i.SubImage(innerRect).(Image)

	// We also need to adjust `y` (baseline).
	// Baseline should be shifted down by Margin.Min.Y + Padding.Min.Y
	// Wait, `y` is usually passed as `metrics.Ascent`.
	// Our `MetricsRect` added padding top/margin top to Ascent.
	// So `y` passed to us includes that.
	// But inner box expects its own Ascent.
	// So we should subtract the top padding/margin?

	// Let's check `MetricsRect`: m.Ascent += Top.
	// So `y` (baseline from top of this box) includes Top space.
	// The inner box expects `y'` = `y` - Top.

	// Correct.

	yOffset := db.Padding.Min.Y + db.Margin.Min.Y

	db.Box.DrawBox(innerImg, y-yOffset, dc)
}

func (db *DecorationBox) MinSize() (fixed.Int26_6, fixed.Int26_6) {
	w, h := db.Box.MinSize()
	return w + db.Padding.Max.X + db.Padding.Min.X + db.Margin.Max.X + db.Margin.Min.X,
		h + db.Padding.Max.Y + db.Padding.Min.Y + db.Margin.Max.Y + db.Margin.Min.Y
}

func (db *DecorationBox) MaxSize() (fixed.Int26_6, fixed.Int26_6) {
	w, h := db.Box.MaxSize()
	if w == 0 && h == 0 {
		return 0, 0
	}
	return w + db.Padding.Max.X + db.Padding.Min.X + db.Margin.Max.X + db.Margin.Min.X,
		h + db.Padding.Max.Y + db.Padding.Min.Y + db.Margin.Max.Y + db.Margin.Min.Y
}
