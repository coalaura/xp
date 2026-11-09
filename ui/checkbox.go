package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

// CheckboxBounds returns the glyph bounds; the label starts four logical pixels later.
func CheckboxBounds(layout theme.Layout, rectangle image.Rectangle) image.Rectangle {
	size := layout.Scale(13)
	top := rectangle.Min.Y + (rectangle.Dy()-size)/2

	return image.Rect(rectangle.Min.X, top, rectangle.Min.X+size, top+size)
}

func drawCheckbox(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control, state State) {
	bounds := CheckboxBounds(layout, rectangle)
	index := 0

	switch {
	case control.Disabled:
		index = 3
	case state.Pressed:
		index = 2
	case state.Hot:
		index = 1
	}

	if control.Mixed && control.Kind == Checkbox {
		index += 8
	} else if control.Checked {
		index += 4
	}

	atlas := checkboxArtwork

	if control.Kind == Radio {
		atlas = radioArtwork
	}

	surface.DrawImage(bounds, atlas, image.Rect(0, index*13, 13, (index+1)*13))
}
