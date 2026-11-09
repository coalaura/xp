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
	one := layout.Scale(1)
	index := 0

	if control.Checked {
		index = 1
	}

	if state.Hot || state.Pressed {
		index = 2
	}

	surface.DrawImage(bounds, checkboxArtwork, image.Rect(0, index*13, 13, (index+1)*13))

	if state.Pressed && !control.Disabled {
		surface.Fill(bounds.Inset(one), 0xC1D2EE)
	}

	if control.Disabled {
		surface.Fill(bounds, 0xCAC8BB)
		surface.Fill(bounds.Inset(one), 0xF5F4EA)
	}

	if !control.Checked || index == 1 && !control.Disabled && !state.Pressed {
		return
	}

	color := uint32(0x21A121)

	if control.Disabled {
		color = 0xA1A192
	}

	// Reuse the captured check mask over the hot, pressed, and disabled surfaces.
	for vertical := range 13 {
		for horizontal := range 13 {
			offset := checkboxArtwork.PixOffset(horizontal, vertical+13)
			if checkboxArtwork.Pix[offset] != 0x21 || checkboxArtwork.Pix[offset+1] != 0xA1 || checkboxArtwork.Pix[offset+2] != 0x21 {
				continue
			}

			// Match DrawImage's nearest-neighbor pixel boundaries at fractional DPI.
			left := bounds.Min.X + (horizontal*bounds.Dx()+12)/13
			top := bounds.Min.Y + (vertical*bounds.Dy()+12)/13
			right := bounds.Min.X + ((horizontal+1)*bounds.Dx()+12)/13
			bottom := bounds.Min.Y + ((vertical+1)*bounds.Dy()+12)/13

			surface.Fill(image.Rect(left, top, right, bottom), color)
		}
	}
}
