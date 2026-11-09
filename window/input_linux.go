package window

import (
	"image"
	"unicode"

	"github.com/coalaura/xp/theme"
	"github.com/coalaura/xp/ui"
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

	// Cursor-font glyphs follow the EWMH resize directions, then arrow, text and hand.
	glyphs := [...]uint16{134, 138, 136, 96, 14, 16, 12, 70, 68, 152, 60}

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

		index := app.controlAt(point)
		if index >= 0 {
			switch app.controls[index].Kind {
			case ui.Input:
				direction = 9
			case ui.Link:
				direction = 10
			}
		}

		if app.comboOpen >= 0 {
			bounds, _ := app.comboBounds()
			if point.In(bounds) {
				direction = 8
			}
		}
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
	point := image.Pt(int(event.EventX), int(event.EventY))

	if event.Detail == 4 || event.Detail == 5 {
		steps := 1

		if event.Detail == 5 {
			steps = -1
		}

		app.controlWheel(point, steps)

		return
	}

	if event.Detail != xproto.ButtonIndex1 {
		return
	}

	app.cancelPress()

	if app.supported[app.atoms["_NET_ACTIVE_WINDOW"]] {
		app.sendRootMessage("_NET_ACTIVE_WINDOW", 1, uint32(event.Time))
	} else {
		xproto.SetInputFocus(app.connection, xproto.InputFocusParent, app.handle, event.Time)
	}

	button := app.layout.ButtonAt(point)
	if button != theme.NoButton {
		app.clearControlFocus()

		app.state.Pressed = button
		app.state.Hot = button
		app.dirty = true

		return
	}

	direction := app.resizeDirection(point)
	if direction >= 0 {
		app.clearControlFocus()
		app.beginDrag(event, direction)

		return
	}

	if point.Y < app.layout.Caption {
		app.clearControlFocus()

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

	app.controlDown(point, event.State&xproto.ModMaskShift != 0)
}

func (app *application) buttonUp(event xproto.ButtonReleaseEvent) {
	if event.Detail != xproto.ButtonIndex1 {
		return
	}

	point := image.Pt(int(event.EventX), int(event.EventY))

	button := app.state.Pressed
	if button == theme.NoButton && !app.drag.active {
		app.controlUp(point)
		app.updateCursor(point)

		return
	}

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
}

func (app *application) pointerMotion(event xproto.MotionNotifyEvent) {
	if app.drag.active {
		app.continueDrag(int(event.RootX), int(event.RootY))

		return
	}

	point := image.Pt(int(event.EventX), int(event.EventY))
	app.controlMove(point)

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
	app.comboDragging = false
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

func (app *application) keySymbol(code xproto.Keycode, state uint16) xproto.Keysym {
	index := (int(code) - app.firstKeycode) * app.keysymsPerKey
	if index < 0 || index >= len(app.keysyms) {
		return 0
	}

	symbol := app.keysyms[index]
	shifted := xproto.Keysym(0)

	if state&xproto.ModMask5 != 0 && app.keysymsPerKey >= 4 && app.keysyms[index+2] != 0 {
		index += 2
		symbol = app.keysyms[index]
	}

	if app.keysymsPerKey > 1 {
		shifted = app.keysyms[index+1]
	}

	if shifted == 0 {
		shifted = xproto.Keysym(unicode.ToUpper(rune(symbol)))
	}

	if state&xproto.ModMaskShift != 0 {
		symbol = shifted
	}

	if state&xproto.ModMask2 != 0 && shifted >= 0xff80 && shifted <= 0xffbd {
		symbol = shifted

		if state&xproto.ModMaskShift != 0 {
			symbol = app.keysyms[index]
		}
	}

	if state&xproto.ModMaskLock != 0 && unicode.IsLetter(rune(symbol)) {
		if state&xproto.ModMaskShift != 0 {
			symbol = xproto.Keysym(unicode.ToLower(rune(symbol)))
		} else {
			symbol = xproto.Keysym(unicode.ToUpper(rune(symbol)))
		}
	}

	return symbol
}

func (app *application) keyDown(event xproto.KeyPressEvent) {
	repeat := app.keys[event.Detail]
	app.keys[event.Detail] = true

	if event.State&xproto.ModMask4 != 0 {
		return
	}

	symbol := app.keySymbol(event.Detail, event.State)

	modifiers := keyModifiers{
		shift:   event.State&xproto.ModMaskShift != 0 || symbol == keyBackTab,
		control: event.State&xproto.ModMaskControl != 0,
		alt:     event.State&xproto.ModMask1 != 0,
	}

	app.controlKeyDown(linuxKey(symbol), modifiers, repeat)

	character := keysymRune(symbol)
	if !modifiers.control && !modifiers.alt && character >= ' ' && !unicode.IsControl(character) {
		app.typeText(string(character))
	}
}

func (app *application) keyUp(event xproto.KeyReleaseEvent) {
	app.keys[event.Detail] = false

	app.controlKeyUp(linuxKey(app.keySymbol(event.Detail, event.State)))
}

func linuxKey(symbol xproto.Keysym) controlKey {
	switch symbol {
	case keyTab, keyBackTab:
		return controlTab
	case keyReturn, keyKPEnter:
		return controlEnter
	case keyEscape:
		return controlEscape
	case 0xff08:
		return controlBackspace
	case 0xffff, 0xff9f:
		return controlDelete
	case 0xff50, 0xff95:
		return controlHome
	case 0xff57, 0xff9c:
		return controlEnd
	case 0xff51, 0xff96:
		return controlLeft
	case 0xff52, 0xff97:
		return controlUp
	case 0xff53, 0xff98:
		return controlRight
	case 0xff54, 0xff99:
		return controlDown
	case 0xff55, 0xff9a:
		return controlPageUp
	case 0xff56, 0xff9b:
		return controlPageDown
	case 0xffc1:
		return controlF4
	}

	return controlKey(unicode.ToLower(keysymRune(symbol)))
}

func keysymRune(symbol xproto.Keysym) rune {
	if symbol >= 0x20 && symbol <= 0x7E || symbol >= 0xA0 && symbol <= 0xFF {
		return rune(symbol)
	}

	if symbol&0xFF000000 == 0x01000000 {
		character := rune(symbol & 0xFFFFFF)
		if character <= unicode.MaxRune && (character < 0xD800 || character > 0xDFFF) {
			return character
		}
	}

	if symbol >= 0xffb0 && symbol <= 0xffb9 {
		return rune('0' + symbol - 0xffb0)
	}

	switch symbol {
	case 0xffaa:
		return '*'
	case 0xffab:
		return '+'
	case 0xffad:
		return '-'
	case 0xffae:
		return '.'
	case 0xffaf:
		return '/'
	}

	return 0
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
