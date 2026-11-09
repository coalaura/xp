package ui

import (
	"image"

	"github.com/coalaura/xp/theme"
)

// SliderThumb returns the same thumb rectangle used for painting and hit testing.
func SliderThumb(layout theme.Layout, rectangle image.Rectangle, control Control) image.Rectangle {
	length := layout.Scale(11)
	width := layout.Scale(21)
	start := rectangle.Min.X + layout.Scale(4)
	travel := max(0, rectangle.Dx()-layout.Scale(8)-length)

	if control.Vertical {
		start = rectangle.Min.Y + layout.Scale(4)
		travel = max(0, rectangle.Dy()-layout.Scale(8)-length)
	}

	span := int64(control.Max) - int64(control.Min)
	position := start

	if span > 0 {
		value := int64(max(control.Min, min(control.Max, control.Value))) - int64(control.Min)
		position += int((value*int64(travel) + span/2) / span)
	}

	if control.Vertical {
		left := rectangle.Min.X + (rectangle.Dx()-width)/2

		return image.Rect(left, position, left+width, position+length)
	}

	top := rectangle.Min.Y + (rectangle.Dy()-width)/2

	return image.Rect(position, top, position+length, top+width)
}

// SliderValue maps a thumb's leading coordinate to a clamped, step-snapped value.
func SliderValue(layout theme.Layout, rectangle image.Rectangle, control Control, position int) int {
	start := rectangle.Min.X + layout.Scale(4)
	travel := rectangle.Dx() - layout.Scale(8) - layout.Scale(11)

	if control.Vertical {
		start = rectangle.Min.Y + layout.Scale(4)
		travel = rectangle.Dy() - layout.Scale(8) - layout.Scale(11)
	}

	if travel <= 0 || control.Max <= control.Min {
		return control.Min
	}

	distance := int64(max(0, min(travel, position-start)))
	span := int64(control.Max) - int64(control.Min)
	step := int64(max(1, control.Step))
	offset := (distance*span + int64(travel)/2) / int64(travel)

	if distance == int64(travel) {
		return control.Max
	}

	offset = (offset + step/2) / step * step

	return int(min(int64(control.Max), int64(control.Min)+offset))
}

func drawSlider(surface theme.Surface, layout theme.Layout, rectangle image.Rectangle, control Control, state State) {
	thumb := SliderThumb(layout, rectangle, control)
	index := interactionIndex(control, state)

	if control.Disabled {
		index = 4
	} else if index == 0 && state.Focused {
		index = 3
	}

	if control.Vertical {
		left := rectangle.Min.X + (rectangle.Dx()-layout.Scale(5))/2
		track := image.Rect(left, rectangle.Min.Y+layout.Scale(9), left+layout.Scale(5), rectangle.Max.Y-layout.Scale(9))

		drawSlices(surface, layout, track, trackVerticalArtwork, trackVerticalArtwork.Bounds(), 2)

		surface.DrawImage(thumb, sliderVerticalArtwork, image.Rect(0, index*11, 21, (index+1)*11))
	} else {
		top := rectangle.Min.Y + (rectangle.Dy()-layout.Scale(5))/2
		track := image.Rect(rectangle.Min.X+layout.Scale(9), top, rectangle.Max.X-layout.Scale(9), top+layout.Scale(5))

		drawSlices(surface, layout, track, trackArtwork, trackArtwork.Bounds(), 2)

		surface.DrawImage(thumb, sliderArtwork, image.Rect(0, index*21, 11, (index+1)*21))
	}

	if state.Focused && !control.Disabled {
		DrawFocus(surface, layout, rectangle.Inset(layout.Scale(1)))
	}
}
