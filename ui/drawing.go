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

func drawReferenceButton(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control, state State) {
	left := min(layout.Scale(7), rectangle.Dx()/2)
	top := min(layout.Scale(7), rectangle.Dy()/2)

	columns := [...]int{rectangle.Min.X, rectangle.Min.X + left, rectangle.Max.X - left, rectangle.Max.X}
	rows := [...]int{rectangle.Min.Y, rectangle.Min.Y + top, rectangle.Max.Y - top, rectangle.Max.Y}

	sourceColumns := [...]int{0, 7, 105, 112}
	sourceRows := [...]int{0, 7, 19, 26}
	index := 0

	switch {
	case control.Disabled:
		index = 3
	case state.Pressed:
		index = 2
	case state.Hot:
		index = 1
	case control.Default:
		index = 4
	}

	offset := index * 26

	// Keep the corners fixed while stretching the text-free edges and center.
	for vertical := range 3 {
		for horizontal := range 3 {
			destination := image.Rect(columns[horizontal], rows[vertical], columns[horizontal+1], rows[vertical+1])
			source := image.Rect(sourceColumns[horizontal], sourceRows[vertical]+offset, sourceColumns[horizontal+1], sourceRows[vertical+1]+offset)

			surface.DrawImage(destination, buttonArtwork, source)
		}
	}
}
