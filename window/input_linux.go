package window

import (
	"image"

	"github.com/coalaura/xp/theme"
	"github.com/jezek/xgb/xproto"
)

const (
	keyTab       = 0xff09
	keyReturn    = 0xff0d
	keyEscape    = 0xff1b
	keyKPEnter   = 0xff8d
	keyBackTab   = 0xfe20
	keySpace     = 0x20
	moveWindow   = 8
	noWindowDrag = -1
)

type windowDrag struct {
	active    bool
	direction int
	rootX     int
	rootY     int
	left      int
	top       int
	width     int
	height    int
}

func (app *application) initializeCursors() {
	fontID, err := xproto.NewFontId(app.connection)
	if err != nil {
		return
	}

	err = xproto.OpenFontChecked(app.connection, fontID, 6, "cursor").Check()
	if err != nil {
		return
	}

	defer xproto.CloseFont(app.connection, fontID)

	// Cursor-font glyphs follow the EWMH resize directions, then the default pointer.
	glyphs := [...]uint16{134, 138, 136, 96, 14, 16, 12, 70, 68}

	for index, glyph := range glyphs {
		cursor, cursorError := xproto.NewCursorId(app.connection)
		if cursorError != nil {
			continue
		}

		cursorError = xproto.CreateGlyphCursorChecked(app.connection, cursor, fontID, fontID, glyph, glyph+1, 0, 0, 0, 65535, 65535, 65535).Check()
		if cursorError == nil {
			app.cursors[index] = cursor
		}
	}
}

func (app *application) updateCursor(point image.Point) {
	direction := app.resizeDirection(point)
	if direction < 0 {
		direction = 8
	}

	if direction != app.cursor {
		app.cursor = direction
		xproto.ChangeWindowAttributes(app.connection, app.handle, xproto.CwCursor, []uint32{uint32(app.cursors[direction])})
	}
}

func (app *application) resizeDirection(point image.Point) int {
	if app.options.Style == Dialog || app.state.Maximized {
		return noWindowDrag
	}

	border := app.layout.Border
	corner := app.layout.Scale(16)

	width := app.layout.Width
	height := app.layout.Height

	left := point.X < border
	right := point.X >= width-border
	top := point.Y < border
	bottom := point.Y >= height-border

	switch {
	case top && point.X < corner || left && point.Y < corner:
		return 0
	case top && point.X >= width-corner || right && point.Y < corner:
		return 2
	case bottom && point.X >= width-corner || right && point.Y >= height-corner:
		return 4
	case bottom && point.X < corner || left && point.Y >= height-corner:
		return 6
	case top:
		return 1
	case right:
		return 3
	case bottom:
		return 5
	case left:
		return 7
	default:
		return noWindowDrag
	}
}

func (app *application) buttonDown(event xproto.ButtonPressEvent) {
	if event.Detail != xproto.ButtonIndex1 {
		return
	}

	app.cancelPress()

	point := image.Pt(int(event.EventX), int(event.EventY))

	if app.supported[app.atoms["_NET_ACTIVE_WINDOW"]] {
		app.sendRootMessage("_NET_ACTIVE_WINDOW", 1, uint32(event.Time))
	} else {
		xproto.SetInputFocus(app.connection, xproto.InputFocusParent, app.handle, event.Time)
	}

	button := app.layout.ButtonAt(point)
	if button != theme.NoButton {
		app.state.Pressed = button
		app.state.Hot = button
		app.dirty = true

		return
	}

	direction := app.resizeDirection(point)
	if direction >= 0 {
		app.beginDrag(event, direction)

		return
	}

	if point.Y < app.layout.Caption {
		distance := app.layout.Scale(4)
		doubleClick := app.lastClick != 0 && uint32(event.Time-app.lastClick) <= 400 && abs(point.X-app.lastClickX) <= distance && abs(point.Y-app.lastClickY) <= distance

		app.lastClick = event.Time
		app.lastClickX = point.X
		app.lastClickY = point.Y

		if doubleClick {
			app.lastClick = 0

			app.toggleMaximize()
		} else {
			app.beginDrag(event, moveWindow)
		}

		return
	}

	app.lastClick = 0

	index := app.controlAt(point)
	if index >= 0 {
		app.controlHot = index
		app.controlPressed = index
		app.controlFocus = index
		app.showFocus = false
		app.dirty = true
	}
}

func (app *application) buttonUp(event xproto.ButtonReleaseEvent) {
	if event.Detail != xproto.ButtonIndex1 {
		return
	}

	point := image.Pt(int(event.EventX), int(event.EventY))
	button := app.state.Pressed
	control := app.controlPressed
	keyboard := app.keyboardPress

	app.cancelPress()
	app.updateCursor(point)

	if button != theme.NoButton && button == app.layout.ButtonAt(point) {
		switch button {
		case theme.Close:
			app.Close()
		case theme.Minimize:
			app.sendRootMessage("WM_CHANGE_STATE", 3)
		case theme.Maximize:
			app.toggleMaximize()
		}
	}

	if !keyboard && control >= 0 && control == app.controlAt(point) {
		app.activateControl(control)
	}
}

func (app *application) pointerMotion(event xproto.MotionNotifyEvent) {
	if app.drag.active {
		app.continueDrag(int(event.RootX), int(event.RootY))

		return
	}

	point := image.Pt(int(event.EventX), int(event.EventY))

	button := app.layout.ButtonAt(point)
	control := app.controlAt(point)

	if button != app.state.Hot || control != app.controlHot {
		app.state.Hot = button
		app.controlHot = control
		app.dirty = true
	}

	app.updateCursor(point)
}

func (app *application) cancelPress() {
	app.state.Pressed = theme.NoButton
	app.controlPressed = -1
	app.keyboardPress = false
	app.drag.active = false
	app.dirty = true
}

func (app *application) beginDrag(event xproto.ButtonPressEvent, direction int) {
	if app.supported[app.atoms["_NET_WM_MOVERESIZE"]] {
		xproto.UngrabPointer(app.connection, event.Time)

		app.sendRootMessage("_NET_WM_MOVERESIZE", uint32(event.RootX), uint32(event.RootY), uint32(direction), 1, 1)

		return
	}

	if app.state.Maximized {
		return
	}

	position, err := xproto.TranslateCoordinates(app.connection, app.handle, app.screen.Root, 0, 0).Reply()
	if err != nil || position == nil {
		return
	}

	// Core X11's implicit button grab keeps the fallback drag alive outside the window.
	app.drag = windowDrag{
		active:    true,
		direction: direction,
		rootX:     int(event.RootX),
		rootY:     int(event.RootY),
		left:      int(position.DstX),
		top:       int(position.DstY),
		width:     app.layout.Width,
		height:    app.layout.Height,
	}
}

func (app *application) continueDrag(rootX, rootY int) {
	drag := &app.drag

	deltaX := rootX - drag.rootX
	deltaY := rootY - drag.rootY

	left := drag.left
	top := drag.top
	width := drag.width
	height := drag.height

	direction := drag.direction

	if direction == moveWindow {
		left += deltaX
		top += deltaY
	} else {
		resizeLeft := direction == 0 || direction == 6 || direction == 7
		resizeRight := direction == 2 || direction == 3 || direction == 4
		resizeTop := direction == 0 || direction == 1 || direction == 2
		resizeBottom := direction == 4 || direction == 5 || direction == 6

		if resizeLeft {
			width -= deltaX
		} else if resizeRight {
			width += deltaX
		}

		if resizeTop {
			height -= deltaY
		} else if resizeBottom {
			height += deltaY
		}

		width = max(app.layout.Scale(app.options.MinWidth), min(width, maxXDimension))
		height = max(app.layout.Scale(app.options.MinHeight), min(height, maxXDimension))

		if resizeLeft {
			left += drag.width - width
		}

		if resizeTop {
			top += drag.height - height
		}
	}

	mask := uint16(xproto.ConfigWindowX | xproto.ConfigWindowY | xproto.ConfigWindowWidth | xproto.ConfigWindowHeight)
	xproto.ConfigureWindow(app.connection, app.handle, mask, []uint32{uint32(left), uint32(top), uint32(width), uint32(height)})
}

func (app *application) loadKeyboard() error {
	setup := xproto.Setup(app.connection)
	count := byte(int(setup.MaxKeycode) - int(setup.MinKeycode) + 1)

	reply, err := xproto.GetKeyboardMapping(app.connection, setup.MinKeycode, count).Reply()
	if err != nil {
		return err
	}

	app.keysyms = reply.Keysyms
	app.keysymsPerKey = int(reply.KeysymsPerKeycode)
	app.firstKeycode = int(setup.MinKeycode)

	return nil
}

func (app *application) keySymbol(code xproto.Keycode) xproto.Keysym {
	index := (int(code) - app.firstKeycode) * app.keysymsPerKey
	if index < 0 || index >= len(app.keysyms) {
		return 0
	}

	return app.keysyms[index]
}

func (app *application) keyDown(event xproto.KeyPressEvent) {
	if app.keys[event.Detail] {
		return
	}

	app.keys[event.Detail] = true

	if event.State&(xproto.ModMaskControl|xproto.ModMask1|xproto.ModMask4) != 0 {
		return
	}

	symbol := app.keySymbol(event.Detail)

	switch symbol {
	case keyTab, keyBackTab:
		app.focusNext(event.State&xproto.ModMaskShift != 0 || symbol == keyBackTab)
	case keyReturn, keyKPEnter:
		app.activateDefault(false)
	case keyEscape:
		app.cancelPress()
		app.activateDefault(true)
	case keySpace:
		if app.controlFocus >= 0 && app.controlPressed < 0 {
			app.controlPressed = app.controlFocus
			app.keyboardPress = true
			app.showFocus = true
			app.dirty = true
		}
	}
}

func (app *application) keyUp(event xproto.KeyReleaseEvent) {
	app.keys[event.Detail] = false

	if app.keySymbol(event.Detail) == keySpace && app.keyboardPress {
		index := app.controlPressed

		app.cancelPress()
		app.activateControl(index)
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
