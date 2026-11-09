// Package theme draws the blue Windows XP Luna frame without platform APIs.
package theme

import "image"

const (
	Background        uint32 = 0xECE9D8
	TitleText         uint32 = 0xFFFFFF
	InactiveTitleText uint32 = 0xD8E4F8
)

const (
	NoButton Button = iota
	Minimize
	Maximize
	Close
)

// Button identifies a caption button.
type Button uint8

// Surface holds top-down, native-endian 0x00RRGGBB pixels. Stride is in pixels.
// The Windows backend draws directly into a DIB; other backends can supply their
// own storage without depending on Win32 or allocating an intermediate image.
type Surface struct {
	Pixels  []uint32
	Width   int
	Height  int
	Stride  int
	clip    image.Rectangle
	clipped bool
}

// State contains the transient appearance of a window frame.
type State struct {
	Active    bool
	Maximized bool
	Hot       Button
	Pressed   Button
}

// Layout is shared by painting and hit testing. Coordinates are physical pixels.
type Layout struct {
	Dialog    bool
	Width     int
	Height    int
	DPI       int
	Border    int
	Caption   int
	Content   image.Rectangle
	Icon      image.Rectangle
	Title     image.Rectangle
	Buttons   [4]image.Rectangle
	maximized bool
}

type gradientStop struct {
	position int
	color    uint32
}

var (
	activeCaption = [...]gradientStop{
		{position: 0, color: 0x0058EE},
		{position: 1, color: 0x3F97FF},
		{position: 2, color: 0x2B90FF},
		{position: 3, color: 0x0372FF},
		{position: 4, color: 0x0365F1},
		{position: 5, color: 0x005CE9},
		{position: 6, color: 0x0058E6},
		{position: 7, color: 0x0054E3},
		{position: 9, color: 0x0054E3},
		{position: 10, color: 0x0055E5},
		{position: 14, color: 0x0055E5},
		{position: 15, color: 0x0055EA},
		{position: 16, color: 0x0055EA},
		{position: 17, color: 0x0058EE},
		{position: 18, color: 0x005BF2},
		{position: 19, color: 0x005AF5},
		{position: 20, color: 0x0060F9},
		{position: 21, color: 0x0164F9},
		{position: 22, color: 0x026AFE},
		{position: 25, color: 0x026AFE},
		{position: 26, color: 0x0065FD},
		{position: 27, color: 0x0060FC},
		{position: 28, color: 0x004DE3},
		{position: 29, color: 0x0043CF},
	}

	inactiveCaption = [...]gradientStop{
		{position: 0, color: 0x688DE0},
		{position: 1, color: 0x98B2E8},
		{position: 2, color: 0x9DB9EB},
		{position: 3, color: 0x8AACE7},
		{position: 4, color: 0x7EA2E4},
		{position: 5, color: 0x7B9DE1},
		{position: 6, color: 0x7A99E0},
		{position: 7, color: 0x7996DE},
		{position: 9, color: 0x7996DE},
		{position: 10, color: 0x7A96DF},
		{position: 12, color: 0x7A96DF},
		{position: 13, color: 0x7A96E0},
		{position: 14, color: 0x7B97E0},
		{position: 15, color: 0x7B99E1},
		{position: 16, color: 0x7B99E1},
		{position: 17, color: 0x7D9BE3},
		{position: 18, color: 0x7D9BE3},
		{position: 19, color: 0x7F9FE6},
		{position: 20, color: 0x7FA2E6},
		{position: 21, color: 0x80A5E7},
		{position: 22, color: 0x81A7E8},
		{position: 23, color: 0x82A9E9},
		{position: 24, color: 0x82A9E9},
		{position: 25, color: 0x81A7E8},
		{position: 26, color: 0x80A5E7},
		{position: 27, color: 0x7E9DE5},
		{position: 28, color: 0x7A93DF},
		{position: 29, color: 0x7A93DF},
	}

	activeMaximizedCaption   = [...]uint32{0x3E96FF, 0x2C8AFF, 0x0370FF, 0x0465F1, 0x015CE9, 0x005AE7, 0x0055E3, 0x0055E3, 0x0055E3, 0x0055E6, 0x0055E6, 0x0056E8, 0x0054EC, 0x0055F1, 0x005AF2, 0x015DF4, 0x005DF9, 0x0162FA, 0x0269FE, 0x036CFE, 0x036CFE, 0x0269FE, 0x0266FB, 0x0162FA, 0x014DE1, 0x0043CF}
	inactiveMaximizedCaption = [...]uint32{0x99B3E9, 0x9DB9EB, 0x88AEE9, 0x7EA1E5, 0x7B9CE1, 0x7998DF, 0x7998DF, 0x7A97DF, 0x7A97DF, 0x7A97E0, 0x7A97E0, 0x7B98E1, 0x7C99E2, 0x7C99E2, 0x7E9BE5, 0x7E9EE5, 0x7FA1E6, 0x80A4E7, 0x81A8E9, 0x81A9E8, 0x82ABE9, 0x81A9E8, 0x80A6E8, 0x7E9EE5, 0x7A95DF, 0x7A95DF}
)

// Fill paints the clipped rectangle with a 0x00RRGGBB color.
func (surface Surface) Fill(rectangle image.Rectangle, color uint32) {
	rectangle = rectangle.Intersect(image.Rect(0, 0, surface.Width, surface.Height))

	if surface.clipped {
		rectangle = rectangle.Intersect(surface.clip)
	}

	for vertical := rectangle.Min.Y; vertical < rectangle.Max.Y; vertical++ {
		row := surface.Pixels[vertical*surface.Stride+rectangle.Min.X : vertical*surface.Stride+rectangle.Max.X]

		for horizontal := range row {
			row[horizontal] = color
		}
	}
}

// Clipped returns a view restricted to the given rectangle without copying pixels.
func (surface Surface) Clipped(rectangle image.Rectangle) Surface {
	if surface.clipped {
		rectangle = rectangle.Intersect(surface.clip)
	}

	surface.clip = rectangle
	surface.clipped = true

	return surface
}

// Scale converts a logical pixel dimension to the layout's DPI.
func (layout Layout) Scale(value int) int {
	return (value*layout.DPI + 48) / 96
}

// ButtonAt returns the caption button under a point, or NoButton.
func (layout Layout) ButtonAt(point image.Point) Button {
	for button := Minimize; button <= Close; button++ {
		if point.In(layout.Buttons[button]) {
			return button
		}
	}

	return NoButton
}

// Maximized removes the sizing borders and fits the shorter maximized caption.
func (layout Layout) Maximized() Layout {
	if layout.Dialog || layout.maximized {
		return layout
	}

	layout.maximized = true
	layout.Border = 0
	layout.Caption = layout.Scale(26)
	layout.Content = image.Rect(0, layout.Caption, layout.Width, layout.Height)

	offset := layout.Scale(4)

	layout.Title = layout.Title.Add(image.Pt(-offset, -offset))
	layout.Title.Max.X += 2 * offset
	layout.Icon = layout.Icon.Add(image.Pt(-offset, -offset))

	for button := Minimize; button <= Close; button++ {
		layout.Buttons[button] = layout.Buttons[button].Add(image.Pt(offset, -offset))
	}

	return layout
}

// NewLayout computes both the visible frame and its interactive geometry.
func NewLayout(width, height, dpi int, dialog bool) Layout {
	if dpi <= 0 {
		dpi = 96
	}

	layout := Layout{
		Dialog: dialog,
		Width:  width,
		Height: height,
		DPI:    dpi,
	}

	layout.Border = layout.Scale(4)
	layout.Caption = layout.Scale(30)

	buttonTop := 6
	titleLeft := 26
	buttonMargin := 6
	lastButton := Minimize

	if dialog {
		layout.Border = layout.Scale(3)
		layout.Caption = layout.Scale(29)
		buttonTop = 5
		titleLeft = 6
		buttonMargin = 5
		lastButton = Close
	}

	layout.Content = image.Rect(layout.Border, layout.Caption, width-layout.Border, height-layout.Border)

	if !dialog {
		layout.Icon = image.Rect(layout.Scale(6), layout.Scale(8), layout.Scale(22), layout.Scale(24))
	}

	right := width - layout.Scale(buttonMargin)

	for button := Close; button >= lastButton; button-- {
		layout.Buttons[button] = image.Rect(right-layout.Scale(21), layout.Scale(buttonTop), right, layout.Scale(buttonTop+21))
		right -= layout.Scale(23)
	}

	layout.Title = image.Rect(layout.Scale(titleLeft), layout.Scale(buttonTop-1), right-layout.Scale(3), layout.Scale(buttonTop+22))

	return layout
}

// Draw paints the entire frame and empty client area. Text is left to the
// platform's font rasterizer using Layout.Title and the caption text colors.
func Draw(surface Surface, layout Layout, state State) {
	surface.Fill(image.Rect(0, 0, surface.Width, surface.Height), Background)

	stops := activeCaption[:]
	outer := uint32(0x00138C)
	inner := uint32(0x0855DD)
	highlight := uint32(0x3D95FF)

	if !state.Active {
		stops = inactiveCaption[:]
		outer = 0x4F53BC
		inner = 0x758CDC
		highlight = 0x98B2E8
	}

	for vertical := 0; vertical < layout.Caption; vertical++ {
		logical := min(29, vertical*96/layout.DPI)
		if layout.Dialog && logical >= 14 {
			// The 29-pixel dialog caption omits one row of the flat blue band.
			logical++
		}

		color := gradient(stops, logical)

		if layout.maximized {
			row := min(25, vertical*96/layout.DPI)
			color = activeMaximizedCaption[row]

			if !state.Active {
				color = inactiveMaximizedCaption[row]
			}
		}

		surface.Fill(image.Rect(0, vertical, layout.Width, vertical+1), color)
	}

	one := layout.Scale(1)

	if !layout.maximized {
		drawBorders(surface, layout, state.Active)

		surface.Fill(image.Rect(0, layout.Scale(8), one, layout.Caption), outer)
		surface.Fill(image.Rect(layout.Width-one, layout.Scale(8), layout.Width, layout.Caption), outer)

		// Stepped corner highlights match Luna's small, pixel-aligned top radius.
		for step := range 5 {
			inset := [...]int{5, 3, 2, 1, 1}[step]

			vertical := layout.Scale(step)
			bottom := layout.Scale(step + 1)

			surface.Fill(image.Rect(layout.Scale(inset), vertical, layout.Scale(inset+1), bottom), highlight)
			surface.Fill(image.Rect(layout.Width-layout.Scale(inset+1), vertical, layout.Width-layout.Scale(inset), bottom), inner)
		}
	}

	if !layout.Icon.Empty() {
		drawIcon(surface, layout)
	}

	for button := Minimize; button <= Close; button++ {
		if !layout.Buttons[button].Empty() {
			drawButton(surface, layout, state, button)
		}
	}
}

func drawBorders(surface Surface, layout Layout, active bool) {
	atlas := frameArtwork
	bottomAtlas := bottomArtwork

	var (
		top       int
		bottomTop int
	)

	height := 570
	border := 4

	if !active {
		top = height
		bottomTop = border
	}

	if layout.Dialog && active {
		atlas = dialogArtwork
		bottomAtlas = dialogBottomArtwork
		height = 161
		border = 3
	}

	bottom := top + height

	leftSource := image.Rect(0, top, border, bottom-border)
	rightSource := image.Rect(border+1, top, 2*border+1, bottom-border)

	if layout.Dialog && !active {
		leftSource.Min.X++
		rightSource.Max.X--
	}

	surface.DrawImage(image.Rect(0, layout.Caption, layout.Border, layout.Height-layout.Border), atlas, leftSource)
	surface.DrawImage(image.Rect(layout.Width-layout.Border, layout.Caption, layout.Width, layout.Height-layout.Border), atlas, rightSource)

	width := bottomAtlas.Rect.Dx()

	surface.DrawImage(image.Rect(layout.Border, layout.Height-layout.Border, layout.Width-layout.Border, layout.Height), bottomAtlas, image.Rect(border, bottomTop, width-border, bottomTop+border))
	surface.DrawImage(image.Rect(0, layout.Height-layout.Border, layout.Border, layout.Height), bottomAtlas, image.Rect(0, bottomTop, border, bottomTop+border))
	surface.DrawImage(image.Rect(layout.Width-layout.Border, layout.Height-layout.Border, layout.Width, layout.Height), bottomAtlas, image.Rect(width-border, bottomTop, width, bottomTop+border))
}

func drawIcon(surface Surface, layout Layout) {
	left := layout.Icon.Min.X
	top := layout.Icon.Min.Y

	fill := func(x, y, width, height int, color uint32) {
		surface.Fill(image.Rect(left+layout.Scale(x), top+layout.Scale(y), left+layout.Scale(x+width), top+layout.Scale(y+height)), color)
	}

	fill(0, 1, 16, 14, 0x173A80)
	fill(1, 2, 14, 12, 0xFFFFFF)
	fill(1, 2, 14, 3, 0x3487EE)
	fill(2, 6, 12, 7, 0xECE9D8)
	fill(3, 7, 3, 5, 0x6C9BE3)
	fill(7, 7, 6, 1, 0xACA899)
	fill(7, 10, 5, 1, 0xACA899)
}

func drawButton(surface Surface, layout Layout, state State, button Button) {
	rectangle := layout.Buttons[button]
	sprite := 0

	switch button {
	case Maximize:
		sprite = 3

		if state.Maximized {
			sprite = 6
		}
	case Close:
		sprite = 9
	}

	if state.Hot == button {
		sprite++

		if state.Pressed == button {
			sprite++
		}
	}

	atlas := captionArtwork

	if !state.Active {
		atlas = inactiveArtwork
		sprite /= 3

		if layout.maximized {
			atlas = inactiveMaximizedArtwork
			sprite = int(button) - 1
		}
	}

	surface.DrawImage(rectangle, atlas, image.Rect(0, sprite*21, 21, (sprite+1)*21))
}

func gradient(stops []gradientStop, position int) uint32 {
	for index := 1; index < len(stops); index++ {
		if position <= stops[index].position {
			previous := stops[index-1]
			next := stops[index]
			amount := (position - previous.position) * 255 / (next.position - previous.position)

			return blend(previous.color, next.color, amount)
		}
	}

	return stops[len(stops)-1].color
}

func blend(first, second uint32, amount int) uint32 {
	weight := uint32(max(0, min(255, amount)))
	inverse := 255 - weight

	red := ((first>>16&255)*inverse + (second>>16&255)*weight) / 255
	green := ((first>>8&255)*inverse + (second>>8&255)*weight) / 255
	blue := ((first&255)*inverse + (second&255)*weight) / 255

	return red<<16 | green<<8 | blue
}
