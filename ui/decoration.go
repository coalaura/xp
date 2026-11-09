package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

func drawDecoration(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control) {
	one := layout.Scale(1)

	if control.Kind == Separator {
		if control.Vertical {
			surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y, rectangle.Min.X+one, rectangle.Max.Y), 0xACA899)
			surface.Fill(image.Rect(rectangle.Min.X+one, rectangle.Min.Y, rectangle.Min.X+2*one, rectangle.Max.Y), 0xFFFFFF)
		} else {
			surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y, rectangle.Max.X, rectangle.Min.Y+one), 0xACA899)
			surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y+one, rectangle.Max.X, rectangle.Min.Y+2*one), 0xFFFFFF)
		}

		return
	}

	if control.Kind == Group {
		index := 0

		if control.Disabled {
			index = 1
		}

		rectangle.Min.Y += layout.Scale(6)

		drawSlices(surface, layout, rectangle, groupArtwork, image.Rect(0, index*28, 112, (index+1)*28), 3)

		return
	}

	if control.Kind == RaisedEdge {
		drawBevel(surface, rectangle, one, 0xF1EFE2, 0x716F64)
		drawBevel(surface, rectangle.Inset(one), one, 0xFFFFFF, 0xACA899)
	} else {
		drawBevel(surface, rectangle, one, 0xACA899, 0xFFFFFF)
		drawBevel(surface, rectangle.Inset(one), one, 0x716F64, 0xF1EFE2)
	}
}

func drawBevel(surface theme.Surface, rectangle image.Rectangle, width int, leading, trailing uint32) {
	if rectangle.Empty() {
		return
	}

	surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y, rectangle.Max.X-width, rectangle.Min.Y+width), leading)
	surface.Fill(image.Rect(rectangle.Min.X, rectangle.Min.Y, rectangle.Min.X+width, rectangle.Max.Y-width), leading)
	surface.Fill(image.Rect(rectangle.Min.X, rectangle.Max.Y-width, rectangle.Max.X, rectangle.Max.Y), trailing)
	surface.Fill(image.Rect(rectangle.Max.X-width, rectangle.Min.Y, rectangle.Max.X, rectangle.Max.Y), trailing)
}

func drawSlices(surface theme.Surface, layout theme.Layout, destination image.Rectangle, atlas *image.NRGBA, source image.Rectangle, border int) {
	if destination.Empty() {
		return
	}

	width := min(layout.Scale(border), destination.Dx()/2)
	height := min(layout.Scale(border), destination.Dy()/2)

	columns := [...]int{destination.Min.X, destination.Min.X + width, destination.Max.X - width, destination.Max.X}
	rows := [...]int{destination.Min.Y, destination.Min.Y + height, destination.Max.Y - height, destination.Max.Y}

	sourceColumns := [...]int{source.Min.X, source.Min.X + border, source.Max.X - border, source.Max.X}
	sourceRows := [...]int{source.Min.Y, source.Min.Y + border, source.Max.Y - border, source.Max.Y}

	for vertical := range 3 {
		for horizontal := range 3 {
			target := image.Rect(columns[horizontal], rows[vertical], columns[horizontal+1], rows[vertical+1])
			crop := image.Rect(sourceColumns[horizontal], sourceRows[vertical], sourceColumns[horizontal+1], sourceRows[vertical+1])

			surface.DrawImage(target, atlas, crop)
		}
	}
}
