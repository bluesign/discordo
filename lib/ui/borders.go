package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/bluesign/discordo/config"
)

const (
	BORDER_LEFT   = 1 << iota
	BORDER_TOP    = 1 << iota
	BORDER_RIGHT  = 1 << iota
	BORDER_BOTTOM = 1 << iota
)

type Bordered struct {
	Invalidatable
	borders      uint
	content      Drawable
	onInvalidate func(d Drawable)
	uiConfig     config.UIConfig
	invisible    bool
}

func NewBordered(
	content Drawable, borders uint, uiConfig config.UIConfig, invisible bool) *Bordered {
	b := &Bordered{
		borders:   borders,
		content:   content,
		uiConfig:  uiConfig,
		invisible: invisible,
	}
	content.OnInvalidate(b.contentInvalidated)
	return b
}

func (bordered *Bordered) Focus(f bool) {
	if di, ok := bordered.content.(DrawableInteractive); ok {
		di.Focus(f)
	}
}

func (bordered *Bordered) contentInvalidated(d Drawable) {
	bordered.Invalidate()
}

func (bordered *Bordered) Children() []Drawable {
	return []Drawable{bordered.content}
}

func (bordered *Bordered) Invalidate() {
	bordered.DoInvalidate(bordered)
}

func (bordered *Bordered) Draw(ctx *Context) {
	x := 0
	y := 0
	width := ctx.Width()
	height := ctx.Height()
	style := bordered.uiConfig.GetStyle(config.STYLE_TREE)

	ctx.Fill(0, 0, width, height, ' ', style)

	if bordered.borders&BORDER_LEFT != 0 {
		if !bordered.invisible {
			ctx.Fill(0, 0, 1, ctx.Height(), '┃', style)
		}
		x += 1
		width -= 1
	}
	if bordered.borders&BORDER_TOP != 0 {
		if !bordered.invisible {
			ctx.Fill(0, 0, ctx.Width(), 1, '━', style)
		}
		y += 1
		height -= 1
	}
	if bordered.borders&BORDER_RIGHT != 0 {
		if !bordered.invisible {
			ctx.Fill(ctx.Width()-1, 0, 1, ctx.Height(), '┃', style)
		}
		width -= 1
	}
	if bordered.borders&BORDER_BOTTOM != 0 {
		if !bordered.invisible {
			ctx.Fill(0, ctx.Height()-1, ctx.Width(), 1, '━', style)
		}
		height -= 1
	}
	subctx := ctx.Subcontext(x, y, width, height)
	bordered.content.Draw(subctx)
}

func (bordered *Bordered) MouseEvent(localX int, localY int, event tcell.Event) {
	switch content := bordered.content.(type) {
	case Mouseable:
		content.MouseEvent(localX, localY, event)
	}
}
func (bordered *Bordered) Event(event tcell.Event) bool {
	switch content := bordered.content.(type) {
	case Interactive:
		return content.Event(event)
	}
	return false
}

// Borders defines various borders used when primitives are drawn.
// These may be changed to accommodate a different look and feel.
var Borders = struct {
	Horizontal  rune
	Vertical    rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune

	LeftT   rune
	RightT  rune
	TopT    rune
	BottomT rune
	Cross   rune

	HorizontalFocus  rune
	VerticalFocus    rune
	TopLeftFocus     rune
	TopRightFocus    rune
	BottomLeftFocus  rune
	BottomRightFocus rune
}{
	Horizontal:  BoxDrawingsLightHorizontal,
	Vertical:    BoxDrawingsLightVertical,
	TopLeft:     BoxDrawingsLightDownAndRight,
	TopRight:    BoxDrawingsLightDownAndLeft,
	BottomLeft:  BoxDrawingsLightUpAndRight,
	BottomRight: BoxDrawingsLightUpAndLeft,

	LeftT:   BoxDrawingsLightVerticalAndRight,
	RightT:  BoxDrawingsLightVerticalAndLeft,
	TopT:    BoxDrawingsLightDownAndHorizontal,
	BottomT: BoxDrawingsLightUpAndHorizontal,
	Cross:   BoxDrawingsLightVerticalAndHorizontal,

	HorizontalFocus:  BoxDrawingsDoubleHorizontal,
	VerticalFocus:    BoxDrawingsDoubleVertical,
	TopLeftFocus:     BoxDrawingsDoubleDownAndRight,
	TopRightFocus:    BoxDrawingsDoubleDownAndLeft,
	BottomLeftFocus:  BoxDrawingsDoubleUpAndRight,
	BottomRightFocus: BoxDrawingsDoubleUpAndLeft,
}
