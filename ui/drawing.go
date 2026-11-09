package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

// DrawFocus draws the dotted keyboard-focus rectangle around a control or label.
func DrawFocus(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle) {
	if rectangle.Empty() {
		return
	}

	one := layout.Scale(1)

	for horizontal := rectangle.Min.X; horizontal < rectangle.Max.X; horizontal += 2 * one {
		surface.Fill(image.Rect(horizontal, rectangle.Min.Y, min(horizontal+one, rectangle.Max.X), rectangle.Min.Y+one), 0x000000)
		surface.Fill(image.Rect(horizontal, rectangle.Max.Y-one, min(horizontal+one, rectangle.Max.X), rectangle.Max.Y), 0x000000)
	}

	for vertical := rectangle.Min.Y; vertical < rectangle.Max.Y; vertical += 2 * one {
		surface.Fill(image.Rect(rectangle.Min.X, vertical, rectangle.Min.X+one, min(vertical+one, rectangle.Max.Y)), 0x000000)
		surface.Fill(image.Rect(rectangle.Max.X-one, vertical, rectangle.Max.X, min(vertical+one, rectangle.Max.Y)), 0x000000)
	}
}

func drawReferenceButton(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, hot bool) {
	left := min(layout.Scale(4), rectangle.Dx()/2)
	top := min(layout.Scale(3), rectangle.Dy()/2)

	columns := [...]int{rectangle.Min.X, rectangle.Min.X + left, rectangle.Max.X - left, rectangle.Max.X}
	rows := [...]int{rectangle.Min.Y, rectangle.Min.Y + top, rectangle.Max.Y - top, rectangle.Max.Y}

	sourceColumns := [...]int{0, 4, 5, 9}
	sourceRows := [...]int{0, 3, 18, 21}
	offset := 0

	if hot {
		offset = 21
	}

	// Keep the corners fixed while stretching the text-free edges and center.
	for vertical := range 3 {
		for horizontal := range 3 {
			destination := image.Rect(columns[horizontal], rows[vertical], columns[horizontal+1], rows[vertical+1])
			source := image.Rect(sourceColumns[horizontal], sourceRows[vertical]+offset, sourceColumns[horizontal+1], sourceRows[vertical+1]+offset)

			surface.DrawImage(destination, buttonArtwork, source)
		}
	}
}
