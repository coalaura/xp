package window

import (
	"image"
	"unsafe"

	"github.com/coalaura/xp/theme"
)

func (app *application) hitTest(position uintptr) uintptr {
	point := nativePoint{
		X: int32(int16(position & 0xFFFF)),
		Y: int32(int16(position >> 16 & 0xFFFF)),
	}

	screenToClient.Call(app.handle, uintptr(unsafe.Pointer(&point)))

	cursor := image.Pt(int(point.X), int(point.Y))

	layout := app.layout
	if !app.state.Maximized && !layout.Dialog {
		border := layout.Border
		corner := layout.Scale(16)

		left := cursor.X < border
		right := cursor.X >= layout.Width-border
		top := cursor.Y < border
		bottom := cursor.Y >= layout.Height-border

		nearLeft := cursor.X < corner
		nearRight := cursor.X >= layout.Width-corner
		nearTop := cursor.Y < corner
		nearBottom := cursor.Y >= layout.Height-corner

		switch {
		case top && nearLeft || left && nearTop:
			return hitTopLeft
		case top && nearRight || right && nearTop:
			return hitTopRight
		case bottom && nearLeft || left && nearBottom:
			return hitBottomLeft
		case bottom && nearRight || right && nearBottom:
			return hitBottomRight
		case left:
			return hitLeft
		case right:
			return hitRight
		case top:
			return hitTop
		case bottom:
			return hitBottom
		}
	}

	if layout.ButtonAt(cursor) != theme.NoButton {
		return hitClient
	}

	if cursor.In(layout.Icon) {
		return hitSystemMenu
	}

	if cursor.Y < layout.Caption {
		return hitCaption
	}

	return hitClient
}

func (app *application) mouseMove(position uintptr) {
	if !app.tracking {
		tracking := trackMouseEvent{
			Size:   uint32(unsafe.Sizeof(trackMouseEvent{})),
			Flags:  2,
			Window: app.handle,
		}

		result, _, _ := trackMouse.Call(uintptr(unsafe.Pointer(&tracking)))
		app.tracking = result != 0
	}

	hot := app.layout.ButtonAt(mousePoint(position))
	if hot != app.state.Hot {
		app.state.Hot = hot

		app.invalidate()
	}

	control := app.controlAt(mousePoint(position))
	if control != app.controlHot {
		app.controlHot = control

		app.invalidate()
	}
}

func (app *application) mouseDown(position uintptr) {
	button := app.layout.ButtonAt(mousePoint(position))
	if button == theme.NoButton {
		app.controlDown(mousePoint(position))

		return
	}

	app.state.Pressed = button
	app.state.Hot = button

	setCapture.Call(app.handle)

	app.invalidate()
}

func (app *application) mouseUp(position uintptr) {
	pressed := app.state.Pressed
	if pressed == theme.NoButton {
		app.controlUp(mousePoint(position))

		return
	}

	hot := app.layout.ButtonAt(mousePoint(position))

	app.cancelPress()

	if hot != pressed {
		return
	}

	command := uintptr(0)

	switch pressed {
	case theme.Minimize:
		command = systemMinimize
	case theme.Maximize:
		command = systemMaximize

		if app.state.Maximized {
			command = systemRestore
		}
	case theme.Close:
		command = systemClose
	}

	postMessage.Call(app.handle, wmSysCommand, command, 0)
}

func (app *application) cancelPress() {
	if app.state.Pressed == theme.NoButton && app.controlPressed == -1 && !app.keyboardPress {
		return
	}

	app.state.Pressed = theme.NoButton
	app.state.Hot = theme.NoButton
	app.controlPressed = -1
	app.keyboardPress = false

	releaseCapture.Call()

	app.invalidate()
}

func mousePoint(position uintptr) image.Point {
	return image.Pt(int(int16(position&0xFFFF)), int(int16(position>>16&0xFFFF)))
}
