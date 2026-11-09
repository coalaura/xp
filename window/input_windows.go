package window

import (
	"image"
	"unicode/utf16"
	"unsafe"

	"github.com/coalaura/xp/theme"
	"github.com/coalaura/xp/ui"
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
	app.controlMove(mousePoint(position))

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
	app.cancelPress()

	button := app.layout.ButtonAt(mousePoint(position))
	if button == theme.NoButton {
		setFocus.Call(app.handle)
		setCapture.Call(app.handle)
		app.controlDown(mousePoint(position), windowsModifiers().shift)

		return
	}

	app.state.Pressed = button
	app.clearControlFocus()
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
	app.state.Pressed = theme.NoButton
	app.state.Hot = theme.NoButton

	app.controlPressed = -1
	app.keyboardPress = false
	app.comboDragging = false

	if app.handle != 0 {
		releaseCapture.Call()
	}

	app.invalidate()
}

func (app *application) keyDown(key, flags uintptr) {
	app.controlKeyDown(windowsKey(key), windowsModifiers(), flags&(1<<30) != 0)
}

func (app *application) keyUp(key uintptr) {
	app.controlKeyUp(windowsKey(key))
}

func (app *application) character(value uint16) {
	if value >= 0xD800 && value <= 0xDBFF {
		app.pendingSurrogate = value

		return
	}

	character := rune(value)

	if value >= 0xDC00 && value <= 0xDFFF {
		if app.pendingSurrogate == 0 {
			return
		}

		character = utf16.DecodeRune(rune(app.pendingSurrogate), rune(value))
	}

	app.pendingSurrogate = 0

	if character >= ' ' && character != 0x7F {
		app.typeText(string(character))
	}
}

func (app *application) clientCursor() uintptr {
	var position nativePoint

	getCursorPos.Call(uintptr(unsafe.Pointer(&position)))
	screenToClient.Call(app.handle, uintptr(unsafe.Pointer(&position)))

	point := image.Pt(int(position.X), int(position.Y))

	if app.comboOpen >= 0 {
		bounds, _ := app.comboBounds()
		if point.In(bounds) {
			return arrowCursor
		}
	}

	index := app.controlAt(point)
	if index >= 0 {
		switch app.controls[index].Kind {
		case ui.Input:
			return textCursor
		case ui.Link:
			return linkCursor
		}
	}

	return arrowCursor
}

func (app *application) mouseWheel(value, position uintptr) {
	point := nativePoint{X: int32(int16(position)), Y: int32(int16(position >> 16))}
	screenToClient.Call(app.handle, uintptr(unsafe.Pointer(&point)))
	app.wheelDelta += int(int16(value >> 16))
	steps := app.wheelDelta / 120
	app.wheelDelta %= 120

	if steps != 0 {
		app.controlWheel(image.Pt(int(point.X), int(point.Y)), steps)
	}
}

func windowsModifiers() keyModifiers {
	shift, _, _ := getKeyState.Call(0x10)
	control, _, _ := getKeyState.Call(0x11)
	alt, _, _ := getKeyState.Call(0x12)

	return keyModifiers{
		shift:   shift&0x8000 != 0,
		control: control&0x8000 != 0,
		alt:     alt&0x8000 != 0,
	}
}

func windowsKey(key uintptr) controlKey {
	switch key {
	case 0x08:
		return controlBackspace
	case 0x09:
		return controlTab
	case 0x0D:
		return controlEnter
	case 0x1B:
		return controlEscape
	case 0x21:
		return controlPageUp
	case 0x22:
		return controlPageDown
	case 0x23:
		return controlEnd
	case 0x24:
		return controlHome
	case 0x25:
		return controlLeft
	case 0x26:
		return controlUp
	case 0x27:
		return controlRight
	case 0x28:
		return controlDown
	case 0x2E:
		return controlDelete
	case 0x73:
		return controlF4
	}

	if key >= 'A' && key <= 'Z' {
		return controlKey(key + 'a' - 'A')
	}

	return controlKey(key)
}

func mousePoint(position uintptr) image.Point {
	return image.Pt(int(int16(position&0xFFFF)), int(int16(position>>16&0xFFFF)))
}
