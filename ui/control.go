// Package ui provides platform-independent, custom-drawn Windows XP controls.
package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

const (
	Label Kind = iota
	Button
	Separator
	Checkbox
)

const (
	Left Anchor = 1 << iota
	Right
	Top
	Bottom
)

// Kind identifies a supported control.
type Kind uint8

// Anchor determines how a control moves or stretches when the window resizes.
// Opposite anchors stretch that axis; no anchors behave like Left | Top.
type Anchor uint8

// Control describes a control in logical, 96-DPI client coordinates. IDs are
// optional, but nonempty IDs must be unique. Bounds excludes the window frame.
// Default and Cancel designate buttons activated by Enter and Escape.
type Control struct {
	ID       string
	Kind     Kind
	Bounds   image.Rectangle
	Anchor   Anchor
	Text     string
	Bold     bool
	Disabled bool
	Default  bool
	Cancel   bool
	Checked  bool
}

// State is the current interactive appearance of a control.
type State struct {
	Hot     bool
	Pressed bool
	Focused bool
}

// Interactive reports whether a control accepts mouse and keyboard input.
func (control Control) Interactive() bool {
	return !control.Disabled && (control.Kind == Button || control.Kind == Checkbox)
}

// Rectangle resolves the control's physical bounds from the original outer
// window dimensions, its anchors, and the current frame layout.
func (control Control) Rectangle(layout theme.Layout, originalWidth, originalHeight int) image.Rectangle {
	content := layout.Content

	rectangle := image.Rect(layout.Scale(control.Bounds.Min.X), layout.Scale(control.Bounds.Min.Y), layout.Scale(control.Bounds.Max.X), layout.Scale(control.Bounds.Max.Y))
	original := theme.NewLayout(layout.Scale(originalWidth), layout.Scale(originalHeight), layout.DPI, layout.Dialog)

	deltaWidth := content.Dx() - original.Content.Dx()
	deltaHeight := content.Dy() - original.Content.Dy()

	if control.Anchor&Right != 0 {
		rectangle.Max.X += deltaWidth

		if control.Anchor&Left == 0 {
			rectangle.Min.X += deltaWidth
		}
	}

	if control.Anchor&Bottom != 0 {
		rectangle.Max.Y += deltaHeight

		if control.Anchor&Top == 0 {
			rectangle.Min.Y += deltaHeight
		}
	}

	return rectangle.Add(content.Min)
}

// Draw paints a control's surface; the backend supplies font rasterization.
func Draw(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control, state State) {
	if rectangle.Empty() {
		return
	}

	one := layout.Scale(1)

	if control.Kind == Separator {
		surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y, rectangle.Max.X, rectangle.Min.Y+one), 0xACA899)
		surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y+one, rectangle.Max.X, rectangle.Min.Y+2*one), 0xFFFFFF)

		return
	}

	if control.Kind == Checkbox {
		drawCheckbox(surface.Clipped(rectangle), layout, rectangle, control, state)

		return
	}

	if control.Kind != Button {
		return
	}

	if !control.Disabled && !state.Pressed && (control.Default || state.Hot || state.Focused) {
		drawReferenceButton(surface, layout, rectangle, state.Hot)

		if state.Focused {
			DrawFocus(surface, layout, rectangle.Inset(layout.Scale(4)))
		}

		return
	}

	border := uint32(0x003C74)
	rim := uint32(0xD6D0C5)
	rimTop := uint32(0xFFFFFF)
	rimInner := uint32(0xECEBE5)

	if control.Default || state.Focused {
		rim = 0x89ADE4
		rimTop = 0xCEE7FF
		rimInner = 0xBCD4F6
	}

	if state.Hot {
		rim = 0xE5A01A
		rimTop = 0xFFF0CF
		rimInner = 0xFBC761
	}

	if state.Pressed {
		rim = 0x9D998D
		rimTop = 0xCDCAC3
		rimInner = 0xE3E3DB
	}

	if control.Disabled {
		border = 0xC9C7BA
		rim = 0xE4E1D5
		rimTop = 0xF1EFE2
		rimInner = 0xF1EFE2
	}

	surface.Fill(image.Rect(rectangle.Min.X+2*one, rectangle.Min.Y, rectangle.Max.X-2*one, rectangle.Max.Y), border)
	surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y+2*one, rectangle.Max.X, rectangle.Max.Y-2*one), border)
	surface.Fill(rectangle.Inset(one), border)

	inside := rectangle.Inset(one)

	for vertical := inside.Min.Y; vertical < inside.Max.Y; vertical++ {
		position := (vertical - inside.Min.Y) * 100 / max(1, inside.Dy()-1)
		color := buttonColor(position, control.Disabled, state.Pressed)

		surface.Fill(image.Rect(inside.Min.X+one, vertical, inside.Max.X-one, vertical+1), color)
	}

	surface.Fill(image.Rect(inside.Min.X, inside.Min.Y+one, inside.Min.X+one, inside.Max.Y-one), rim)
	surface.Fill(image.Rect(inside.Max.X-one, inside.Min.Y+one, inside.Max.X, inside.Max.Y-one), rim)
	surface.Fill(image.Rect(inside.Min.X+one, inside.Max.Y-one, inside.Max.X-one, inside.Max.Y), rim)

	if control.Default || state.Hot || state.Focused || state.Pressed {
		surface.Fill(image.Rect(inside.Min.X+one, inside.Min.Y, inside.Max.X-one, inside.Min.Y+one), rimTop)
		surface.Fill(image.Rect(inside.Min.X+one, inside.Min.Y+one, inside.Min.X+2*one, inside.Max.Y-one), rimInner)
		surface.Fill(image.Rect(inside.Max.X-2*one, inside.Min.Y+one, inside.Max.X-one, inside.Max.Y-one), rimInner)
		surface.Fill(image.Rect(inside.Min.X+one, inside.Max.Y-2*one, inside.Max.X-one, inside.Max.Y-one), rimInner)
	}

	if state.Focused && !control.Disabled {
		DrawFocus(surface, layout, rectangle.Inset(layout.Scale(4)))
	}
}

func buttonColor(position int, disabled, pressed bool) uint32 {
	if disabled {
		return 0xF5F4EA
	}

	if pressed {
		switch {
		case position < 8:
			return mix(0xCDCAC3, 0xE3E3DB, position*255/8)
		case position < 94:
			return mix(0xE3E3DB, 0xE5E5DE, (position-8)*255/86)
		default:
			return mix(0xE5E5DE, 0xF2F2F1, (position-94)*255/6)
		}
	}

	if position < 86 {
		return mix(0xFFFFFF, 0xECEBE5, position*255/86)
	}

	return mix(0xECEBE5, 0xD8D0C4, (position-86)*255/14)
}

func mix(first, second uint32, amount int) uint32 {
	weight := uint32(amount)
	inverse := 255 - weight

	red := ((first>>16&255)*inverse + (second>>16&255)*weight) / 255
	green := ((first>>8&255)*inverse + (second>>8&255)*weight) / 255
	blue := ((first&255)*inverse + (second&255)*weight) / 255

	return red<<16 | green<<8 | blue
}
