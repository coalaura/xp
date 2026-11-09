package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

// FieldTextBounds is the clipped text viewport of an input or combo box.
func FieldTextBounds(layout theme.Layout, rectangle image.Rectangle, kind Kind) image.Rectangle {
	result := rectangle.Inset(layout.Scale(3))
	result.Min.X += layout.Scale(2)

	if kind == ComboBox {
		result.Max.X -= layout.Scale(19)
	}

	return result
}

func drawField(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control, state State) {
	index := 0

	switch {
	case control.Disabled:
		index = 3
	case control.Kind == Input && control.ReadOnly:
		index = 4
	case control.Kind == Input && state.ActiveFocus:
		index = 2
	case control.Kind == Input && state.Hot:
		index = 1
	}

	drawSlices(surface, layout, rectangle, inputArtwork, image.Rect(0, index*23, 111, (index+1)*23), 2)

	if control.Kind != ComboBox {
		return
	}

	index = interactionIndex(control, state)

	arrow := rectangle.Inset(layout.Scale(1))
	arrow.Min.X = max(arrow.Min.X, arrow.Max.X-layout.Scale(17))

	surface.DrawImage(arrow, comboArtwork, image.Rect(0, index*22, 17, (index+1)*22))
}

func interactionIndex(control Control, state State) int {
	switch {
	case control.Disabled:
		return 3
	case state.Pressed:
		return 2
	case state.Hot:
		return 1
	default:
		return 0
	}
}
