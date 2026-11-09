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
	Heading
	Link
	Radio
	Slider
	Input
	ComboBox
	SunkenEdge
	RaisedEdge
	Group
)

// Text is the descriptive name for a label; Label remains supported.
const Text = Label

const (
	Unchecked CheckState = iota
	Checked
	Mixed
)

const (
	Left Anchor = 1 << iota
	Right
	Top
	Bottom
)

// CheckState is the logical state of a checkbox or radio button.
type CheckState uint8

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
	// Mixed and Checked are mutually exclusive. ThreeState cycles through all three states.
	Mixed      bool
	ThreeState bool
	// RadioGroup identifies mutually exclusive radio buttons, independently of visual groups.
	RadioGroup string
	// Vertical applies to sliders and separators. Slider values increase downwards vertically.
	Vertical bool
	// Min, Max, Value and Step describe a slider. Step zero means one.
	Min   int
	Max   int
	Value int
	Step  int
	// Password and ReadOnly apply to single-line inputs.
	Password bool
	ReadOnly bool
	// Items and Selected describe a list-only combo box. Selected -1 means no selection.
	Items    []string
	Selected int
}

// State is the current interactive appearance of a control.
type State struct {
	Hot     bool
	Pressed bool
	Focused bool
	// ActiveFocus includes mouse focus, independently of keyboard focus cues.
	ActiveFocus bool
}

// Interactive reports whether a control accepts mouse and keyboard input.
func (control Control) Interactive() bool {
	if control.Disabled {
		return false
	}

	switch control.Kind {
	case Button, Checkbox, Link, Radio, Slider, Input, ComboBox:
		return true
	default:
		return false
	}
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

	surface = surface.Clipped(rectangle)

	switch control.Kind {
	case Separator, SunkenEdge, RaisedEdge, Group:
		drawDecoration(surface, layout, rectangle, control)
	case Checkbox, Radio:
		drawCheckbox(surface.Clipped(rectangle), layout, rectangle, control, state)
	case Slider:
		drawSlider(surface, layout, rectangle, control, state)
	case Input, ComboBox:
		drawField(surface, layout, rectangle, control, state)
	case Button:
		drawReferenceButton(surface, layout, rectangle, control, state)

		if state.Focused && !control.Disabled {
			DrawFocus(surface, layout, rectangle.Inset(layout.Scale(4)))
		}
	}
}
