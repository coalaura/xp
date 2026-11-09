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

	drawReferenceButton(surface, layout, rectangle, control, state)

	if state.Focused && !control.Disabled {
		DrawFocus(surface, layout, rectangle.Inset(layout.Scale(4)))
	}
}
