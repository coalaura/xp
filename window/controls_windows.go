package window

import (
	"errors"
	"image"
	"slices"
	"syscall"
	"unsafe"

	"github.com/coalaura/xp/ui"
)

type nativeControl struct {
	control ui.Control
	text    []uint16
}

func (app *application) Close() {
	postMessage.Call(app.handle, wmClose, 0, 0)
}

func (app *application) SetText(id, text string) error {
	encoded, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	for index := range app.controls {
		control := &app.controls[index]
		if id != "" && control.control.ID == id {
			control.control.Text = text
			control.text = encoded

			app.invalidate()

			return nil
		}
	}

	return errors.New("xp: no control with that ID")
}

func (app *application) Checked(id string) (bool, error) {
	control, err := app.checkbox(id)
	if err != nil {
		return false, err
	}

	return control.Checked, nil
}

func (app *application) SetChecked(id string, checked bool) error {
	control, err := app.checkbox(id)
	if err != nil {
		return err
	}

	control.Checked = checked

	app.invalidate()

	return nil
}

func (app *application) checkbox(id string) (*ui.Control, error) {
	for index := range app.controls {
		control := &app.controls[index].control
		if id != "" && control.ID == id && control.Kind == ui.Checkbox {
			return control, nil
		}
	}

	return nil, errors.New("xp: no checkbox with that ID")
}

func (app *application) initializeControls() error {
	app.controls = make([]nativeControl, len(app.options.Controls))

	for index := range app.options.Controls {
		control := &app.options.Controls[index]

		text, err := syscall.UTF16FromString(control.Text)
		if err != nil {
			return err
		}

		app.controls[index] = nativeControl{
			control: *control,
			text:    text,
		}
	}

	return nil
}

func (app *application) controlAt(point image.Point) int {
	if !point.In(app.layout.Content) {
		return -1
	}

	for index, v := range slices.Backward(app.controls) {
		control := &v.control
		if !control.Interactive() {
			continue
		}

		rectangle := control.Rectangle(app.layout, app.options.Width, app.options.Height)
		if point.In(rectangle) {
			return index
		}
	}

	return -1
}

func (app *application) controlDown(point image.Point) {
	index := app.controlAt(point)
	if index < 0 {
		return
	}

	app.controlFocus = index
	app.controlPressed = index
	app.controlHot = index
	app.showFocus = false

	setFocus.Call(app.handle)
	setCapture.Call(app.handle)

	app.invalidate()
}

func (app *application) controlUp(point image.Point) {
	pressed := app.controlPressed
	if pressed < 0 {
		return
	}

	hot := app.controlAt(point)

	app.cancelPress()

	if hot == pressed {
		app.activateControl(pressed)
	}
}

func (app *application) activateControl(index int) {
	if index < 0 || index >= len(app.controls) || !app.controls[index].control.Interactive() {
		return
	}

	control := &app.controls[index].control
	if control.Kind == ui.Checkbox {
		control.Checked = !control.Checked

		app.invalidate()
	}

	if app.options.OnClick != nil {
		app.options.OnClick(app, app.controls[index].control.ID)
	}
}

func (app *application) keyDown(key, flags uintptr) {
	if flags&(1<<30) != 0 {
		return
	}

	switch key {
	case 9:
		app.cancelPress()

		shift, _, _ := getKeyState.Call(0x10)
		app.moveFocus(int16(shift) < 0)
	case 32:
		if app.controlFocus >= 0 && app.controls[app.controlFocus].control.Interactive() {
			app.keyboardPress = true
			app.showFocus = true

			app.invalidate()
		}
	case 13:
		if app.controlFocus >= 0 && app.controls[app.controlFocus].control.Kind == ui.Button {
			app.activateControl(app.controlFocus)

			return
		}

		for index := range app.controls {
			if app.controls[index].control.Default {
				app.activateControl(index)

				return
			}
		}
	case 27:
		for index := range app.controls {
			if app.controls[index].control.Cancel {
				app.activateControl(index)

				return
			}
		}
	}
}

func (app *application) keyUp(key uintptr) {
	if key != 32 || !app.keyboardPress {
		return
	}

	app.keyboardPress = false

	app.invalidate()
	app.activateControl(app.controlFocus)
}

func (app *application) moveFocus(backward bool) {
	count := len(app.controls)
	index := app.controlFocus
	direction := 1

	if backward {
		direction = -1

		if index < 0 {
			index = 0
		}
	}

	for range count {
		index = (index + direction + count) % count

		control := &app.controls[index].control
		if control.Interactive() {
			app.controlFocus = index
			app.showFocus = true

			app.invalidate()

			return
		}
	}
}

func (app *application) deleteFonts() {
	if app.font != 0 {
		deleteObject.Call(app.font)
		app.font = 0
	}

	for index, font := range app.controlFonts {
		if font != 0 {
			deleteObject.Call(font)
			app.controlFonts[index] = 0
		}
	}
}

func (app *application) drawControls() error {
	if len(app.controls) == 0 {
		return nil
	}

	dc := app.buffer.dc

	saved, _, lastError := saveDC.Call(dc)
	if saved == 0 {
		return nativeError("SaveDC", lastError)
	}

	defer restoreDC.Call(dc, saved)

	content := app.layout.Content
	intersectClipRect.Call(dc, uintptr(content.Min.X), uintptr(content.Min.Y), uintptr(content.Max.X), uintptr(content.Max.Y))

	surface := app.buffer.surface.Clipped(content)

	for index := range app.controls {
		control := &app.controls[index]
		rectangle := control.control.Rectangle(app.layout, app.options.Width, app.options.Height)

		state := ui.State{
			Hot:     app.controlHot == index,
			Pressed: app.controlPressed == index && app.controlHot == index || app.keyboardPress && app.controlFocus == index,
			Focused: app.controlFocus == index && app.showFocus && app.state.Active,
		}

		// Finish batched text writes before the next control writes directly to the DIB.
		gdiFlush.Call()

		ui.Draw(surface, app.layout, rectangle, control.control, state)

		if control.control.Kind == ui.Separator || len(control.text) <= 1 || rectangle.Empty() {
			continue
		}

		err := app.drawControlText(control, rectangle, state)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *application) drawControlText(control *nativeControl, rectangle image.Rectangle, state ui.State) error {
	fontIndex := 0
	weight := uintptr(400)

	if control.control.Bold {
		fontIndex = 1
		weight = 700
	}

	if app.controlFonts[fontIndex] == 0 {
		height := -int32(app.layout.Scale(11))

		font, _, lastError := createFont.Call(uintptr(height), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 3, 0, uintptr(unsafe.Pointer(controlFace)))
		if font == 0 {
			return nativeError("CreateFontW", lastError)
		}

		app.controlFonts[fontIndex] = font
	}

	selectObject.Call(app.buffer.dc, app.controlFonts[fontIndex])
	setBackgroundMode.Call(app.buffer.dc, 1)

	color := uint32(0x000000)

	if control.control.Disabled {
		color = 0xACA899
	}

	setTextColor.Call(app.buffer.dc, colorRef(color))

	flags := uintptr(0x810) // Literal ampersands and word wrapping.

	if control.control.Kind == ui.Button {
		flags = 0x8825 // Centered, single-line, ellipsized, literal text.

		rectangle = rectangle.Inset(app.layout.Scale(3))

		if state.Pressed {
			rectangle = rectangle.Add(image.Pt(app.layout.Scale(1), app.layout.Scale(1)))
		}
	}

	if control.control.Kind == ui.Checkbox {
		flags = 0x8824 // Left-aligned, vertically centered, single-line literal text.

		glyph := ui.CheckboxBounds(app.layout, rectangle)

		rectangle.Min.X = glyph.Max.X + app.layout.Scale(4)
	}

	if rectangle.Empty() {
		return nil
	}

	bounds := nativeRect{
		Left:   int32(rectangle.Min.X),
		Top:    int32(rectangle.Min.Y),
		Right:  int32(rectangle.Max.X),
		Bottom: int32(rectangle.Max.Y),
	}

	drawText.Call(app.buffer.dc, uintptr(unsafe.Pointer(&control.text[0])), uintptr(len(control.text)-1), uintptr(unsafe.Pointer(&bounds)), flags)

	if control.control.Kind == ui.Checkbox && state.Focused {
		measured := bounds

		drawText.Call(app.buffer.dc, uintptr(unsafe.Pointer(&control.text[0])), uintptr(len(control.text)-1), uintptr(unsafe.Pointer(&measured)), 0xC20)

		height := int(measured.Bottom - measured.Top)
		top := rectangle.Min.Y + (rectangle.Dy()-height)/2

		focus := image.Rect(rectangle.Min.X, top, min(int(measured.Right), rectangle.Max.X), top+height).Inset(-app.layout.Scale(1))

		gdiFlush.Call()

		ui.DrawFocus(app.buffer.surface.Clipped(app.layout.Content), app.layout, focus)
	}

	return nil
}
