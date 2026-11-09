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
		layout.Icon = image.Rect(layout.Scale(5), layout.Scale(8), layout.Scale(21), layout.Scale(24))
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

	atlas := titleArtwork
	height := 30
	center := 26

	if layout.maximized {
		atlas = titleMaximizedArtwork
		height = 24
		center = 0
	}

	offset := 0

	if !state.Active {
		offset = height
	}

	left := 26

	if layout.Dialog {
		left = 4
	}

	leftWidth := min(layout.Scale(left), layout.Width/2)
	rightWidth := min(layout.Scale(6), layout.Width/2)

	for vertical := 0; vertical < layout.Caption; vertical++ {
		logical := min(29, vertical*96/layout.DPI)
		if layout.Dialog && logical >= 14 {
			// The 29-pixel dialog caption omits one row of the flat blue band.
			logical++
		}

		if layout.maximized {
			logical = min(height-1, vertical*height/layout.Caption)
		}

		pixel := atlas.NRGBAAt(center, offset+logical)
		color := uint32(pixel.R)<<16 | uint32(pixel.G)<<8 | uint32(pixel.B)

		surface.Fill(image.Rect(0, vertical, layout.Width, vertical+1), color)

		if !layout.maximized {
			row := offset + logical

			surface.DrawImage(image.Rect(0, vertical, leftWidth, vertical+1), atlas, image.Rect(0, row, left, row+1))
			surface.DrawImage(image.Rect(layout.Width-rightWidth, vertical, layout.Width, vertical+1), atlas, image.Rect(27, row, 33, row+1))
		}
	}

	if !layout.maximized {
		drawBorders(surface, layout, state.Active)
	}

	if layout.maximized && !layout.Icon.Empty() {
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

	border := 4

	if layout.Dialog {
		atlas = dialogArtwork
		bottomAtlas = dialogBottomArtwork
		border = 3
	}

	height := atlas.Rect.Dy() / 2

	var (
		top       int
		bottomTop int
	)

	if !active {
		top = height
		bottomTop = border
	}

	bottom := top + height

	leftSource := image.Rect(0, top, border, bottom-border)
	rightSource := image.Rect(border+1, top, 2*border+1, bottom-border)

	surface.DrawImage(image.Rect(0, layout.Caption, layout.Border, layout.Height-layout.Border), atlas, leftSource)
	surface.DrawImage(image.Rect(layout.Width-layout.Border, layout.Caption, layout.Width, layout.Height-layout.Border), atlas, rightSource)

	width := bottomAtlas.Rect.Dx()

	surface.DrawImage(image.Rect(layout.Border, layout.Height-layout.Border, layout.Width-layout.Border, layout.Height), bottomAtlas, image.Rect(border, bottomTop, width-border, bottomTop+border))
	surface.DrawImage(image.Rect(0, layout.Height-layout.Border, layout.Border, layout.Height), bottomAtlas, image.Rect(0, bottomTop, border, bottomTop+border))
	surface.DrawImage(image.Rect(layout.Width-layout.Border, layout.Height-layout.Border, layout.Width, layout.Height), bottomAtlas, image.Rect(width-border, bottomTop, width, bottomTop+border))
}

func drawIcon(surface Surface, layout Layout) {
	surface.DrawImage(layout.Icon, iconArtwork, iconArtwork.Bounds())
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
	}

	surface.DrawImage(rectangle, atlas, image.Rect(0, sprite*21, 21, (sprite+1)*21))
}

func blend(first, second uint32, amount int) uint32 {
	weight := uint32(max(0, min(255, amount)))
	inverse := 255 - weight

	red := ((first>>16&255)*inverse + (second>>16&255)*weight) / 255
	green := ((first>>8&255)*inverse + (second>>8&255)*weight) / 255
	blue := ((first&255)*inverse + (second&255)*weight) / 255

	return red<<16 | green<<8 | blue
}
